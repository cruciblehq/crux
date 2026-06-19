package recipe

import (
	"context"
	"io"
	"log/slog"
	"maps"
	"path"
	"path/filepath"
	"strings"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/resource/oci"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
)

// Shell used to execute a Run step when neither the step nor the stage
// specifies one.
const defaultShell = "/bin/sh"

// slog attribute value marking a build log line as standard output.
const logStreamStdout = "stdout"

// slog attribute value marking a build log line as standard error.
const logStreamStderr = "stderr"

// Executes a single step, dispatching to the appropriate handler.
//
// Run steps invoke the command in the live container. Copy steps append a tar
// layer constructed from host files or extracted from a prior stage. Standalone
// modifier steps update the persistent stage state and optionally update the
// image configuration. Platform groups recurse over their children when the
// target platform matches.
func (b *Builder) executeStep(ctx context.Context, ctr *compute.Container, step *manifest.Step, state *stageState, stageImages map[string]string) error {
	if step.Platform != "" && len(step.Steps) > 0 {
		return b.executePlatformGroup(ctx, ctr, step, state, stageImages)
	}

	if step.Platform != "" && !matchesBuildPlatform(step.Platform) {
		return nil
	}

	switch {
	case step.Run != "":
		return b.runStep(ctx, ctr, step, state)
	case step.Copy != "":
		return b.copyStep(ctx, ctr, step, state, stageImages)
	default:
		return b.applyModifierStep(ctx, ctr, step, state)
	}
}

// Executes a platform-scoped group of steps when the platform matches.
//
// Group-level modifier fields are applied to the persistent stage state before
// the children are executed. If the platform does not match the current build
// target, all children are skipped.
func (b *Builder) executePlatformGroup(ctx context.Context, ctr *compute.Container, step *manifest.Step, state *stageState, stageImages map[string]string) error {
	if !matchesBuildPlatform(step.Platform) {
		return nil
	}
	if hasModifiers(step) {
		applyStateModifiers(state, step)
	}
	for i := range step.Steps {
		if err := b.executeStep(ctx, ctr, &step.Steps[i], state, stageImages); err != nil {
			return err
		}
	}
	return nil
}

// Executes a Run step inside the live container.
//
// Starts the container, execs the command with the effective shell, environment,
// workdir, and user from the persistent stage state merged with any step-local
// overrides, then stops the container. Filesystem changes accumulate in the
// container's writable snapshot.
func (b *Builder) runStep(ctx context.Context, ctr *compute.Container, step *manifest.Step, state *stageState) error {
	shell := state.shell
	if step.Shell != "" {
		shell = step.Shell
	}
	if shell == "" {
		shell = defaultShell
	}

	workdir := state.workdir
	if step.Workdir != "" {
		workdir = step.Workdir
	}

	user := state.user
	if step.User != "" {
		user = step.User
	}

	if err := ctr.Start(ctx); err != nil {
		return err
	}
	defer ctr.Stop(context.Background())

	env := maps.Clone(state.env)
	maps.Copy(env, step.Env)
	return ctr.Exec(ctx, step.Run, compute.RuntimeOptions{
		Shell:   shell,
		Env:     env,
		Workdir: workdir,
		User:    user,
		Stdout:  newStreamWriter(ctx, slog.LevelInfo, logStreamStdout),
		Stderr:  newStreamWriter(ctx, slog.LevelInfo, logStreamStderr),
	})
}

// Executes a Copy step by appending a tar layer to the container.
//
// For local copies the host source path is resolved relative to the manifest
// directory. For cross-stage copies (src of the form "stageName:/path") the
// source is extracted from the named stage's committed image. The destination
// defaults to the current workdir when set, otherwise to the image root.
func (b *Builder) copyStep(ctx context.Context, ctr *compute.Container, step *manifest.Step, state *stageState, stageImages map[string]string) error {
	src, dest, ok := strings.Cut(step.Copy, " ")
	if !ok {
		return crex.Newf(ErrBuild, "malformed copy step %q, expected \"src dest\"", step.Copy)
	}

	workdir := state.workdir
	if step.Workdir != "" {
		workdir = step.Workdir
	}
	if dest == "." || dest == "" {
		if workdir != "" {
			dest = workdir
		} else {
			dest = "/"
		}
	}

	if stageName, srcPath, found := strings.Cut(src, ":"); found {
		return b.crossStageCopy(ctx, ctr, stageName, srcPath, dest, stageImages)
	}

	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(oci.WriteCopyTar(pw, filepath.Join(b.workdir, src), dest))
	}()

	return ctr.Copy(ctx, pr)
}

// Copies srcPath from the named prior stage into ctr at destPath.
//
// Extracts srcPath directly from the stage's committed image, rewrites the
// tar paths to destPath, and copies the result into ctr.
func (b *Builder) crossStageCopy(ctx context.Context, ctr *compute.Container, stageName, srcPath, destPath string, stageImages map[string]string) error {
	stageRef, ok := stageImages[stageName]
	if !ok {
		return crex.Newf(ErrBuild, "unknown stage %q in cross-stage copy", stageName)
	}

	if !path.IsAbs(srcPath) {
		cfg, err := b.client.Inspect(ctx, stageRef)
		if err != nil {
			return crex.Wrapf(ErrBuild, err, "get config of stage %q", stageName)
		}
		workdir := cfg.WorkingDir
		if workdir == "" {
			workdir = "/"
		}
		srcPath = path.Join(workdir, srcPath)
	}

	rc, err := b.client.Extract(ctx, stageRef, srcPath)
	if err != nil {
		return crex.Wrapf(ErrBuild, err, "extract %q from stage %q", srcPath, stageName)
	}

	pr, pw := io.Pipe()
	go func() {
		err := oci.RewriteTarPaths(pw, rc, srcPath, destPath)
		rc.Close()
		pw.CloseWithError(err)
	}()

	return ctr.Copy(ctx, pr)
}

// Applies a standalone modifier step.
//
// Updates the persistent stage state and, for fields visible in the final
// image (Env, Workdir, User), updates the image configuration as well. Shell
// changes are build-time only and not reflected in the image config.
func (b *Builder) applyModifierStep(ctx context.Context, ctr *compute.Container, step *manifest.Step, state *stageState) error {
	applyStateModifiers(state, step)

	if step.Workdir == "" && step.User == "" && len(step.Env) == 0 {
		return nil
	}

	cfg, err := ctr.Inspect(ctx)
	if err != nil {
		return crex.Wrapf(ErrBuild, err, "read config")
	}
	if step.Workdir != "" {
		cfg.WorkingDir = step.Workdir
	}
	if step.User != "" {
		cfg.User = step.User
	}
	if len(step.Env) > 0 {
		overrides := make([]string, 0, len(step.Env))
		for k, v := range step.Env {
			overrides = append(overrides, k+"="+v)
		}
		cfg.Env = compute.MergeEnv(cfg.Env, overrides)
	}
	ctr.Configure(cfg)
	return nil
}

// Whether a step has any modifier fields set.
func hasModifiers(s *manifest.Step) bool {
	return s.Shell != "" || s.Workdir != "" || s.User != "" || len(s.Env) > 0
}

// Applies a step's modifier fields to the persistent stage state.
func applyStateModifiers(state *stageState, step *manifest.Step) {
	if step.Shell != "" {
		state.shell = step.Shell
	}
	if step.Workdir != "" {
		state.workdir = step.Workdir
	}
	if step.User != "" {
		state.user = step.User
	}
	for k, v := range step.Env {
		if state.env == nil {
			state.env = make(map[string]string)
		}
		state.env[k] = v
	}
}

// Whether the current build target platform matches the given filter.
//
// Only Linux targets are supported. Non-Linux platforms are silently skipped.
func matchesBuildPlatform(platform string) bool {
	// TODO: compare against the actual target platform from the build context.
	targetOS, _, _ := strings.Cut(platform, "/")
	return targetOS == "linux"
}
