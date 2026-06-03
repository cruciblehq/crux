package resource

import (
	"context"
	"io"
	"log/slog"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/manifest"
)

// Persistent mutable build state within a single stage.
//
// Tracks the cumulative effect of standalone modifier steps (shell, workdir,
// user, env) that persist across subsequent steps in the same stage.
type stageState struct {
	shell   string            // Current shell executable, empty means /bin/sh.
	workdir string            // Current working directory override, empty uses image default.
	user    string            // Current user override in uid:gid format, empty uses image default.
	env     map[string]string // Accumulated environment overrides for subsequent steps.
}

// Runs a single stage: compiles affordances into a security policy, imports
// the base image with that policy, and executes each step in order.
//
// stageImages is updated with the final committed image ref under the stage
// name when the stage has a name, making it available as a copy source for
// later stages. The caller is responsible for closing the returned container.
func (r *Builder) runStage(ctx context.Context, num int, stage *manifest.Stage, stageImages map[string]string) (*compute.Container, error) {
	security, err := r.applyGrants(ctx, stage.Grants)
	if err != nil {
		return nil, crex.Wrapf(ErrBuild, "stage %d: compile grants: %w", num, err)
	}

	ctr, err := r.importBase(ctx, stage, compute.RuntimeOptions{OCI: *security})
	if err != nil {
		return nil, crex.Wrapf(ErrBuild, "stage %d: import base: %w", num, err)
	}

	state := &stageState{}
	for j := range stage.Steps {
		if err := r.executeStep(ctx, ctr, &stage.Steps[j], state, stageImages); err != nil {
			ctr.Destroy(ctx)
			return nil, crex.Wrapf(ErrBuild, "stage %d step %d: %w", num, j+1, err)
		}
	}

	if stage.Name != "" {
		img, err := ctr.Commit(ctx, stage.Name)
		if err != nil {
			ctr.Destroy(ctx)
			return nil, crex.Wrapf(ErrBuild, "stage %d: commit: %w", num, err)
		}
		stageImages[stage.Name] = img
	}

	return ctr, nil
}

// Resolves and imports the base image for a stage.
//
// When stage.From is nil the stage starts from scratch and a minimal empty
// OCI image is imported. When From is set, the referenced runtime is pulled
// and its image.tar is imported into the compute backend.
func (r *Builder) importBase(ctx context.Context, stage *manifest.Stage, opts compute.RuntimeOptions) (*compute.Container, error) {
	if stage.From == "" {
		return r.importScratch(ctx, opts)
	}

	ref, err := r.src.Parse(string(manifest.TypeRuntime), stage.From)
	if err != nil {
		return nil, err
	}

	result, err := r.src.Pull(ctx, ref)
	if err != nil {
		return nil, err
	}

	imgPath := filepath.Join(result.Extracted, files.ImageFile)
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, crex.Wrap(ErrFileSystemOperation, err)
	}
	defer f.Close()

	imgRef, err := r.sess.Import(ctx, f)
	if err != nil {
		return nil, err
	}
	return r.sess.Load(ctx, imgRef, opts)
}

// Creates and imports a minimal empty (scratch) OCI image.
//
// The image has no layers and an empty configuration. It is imported into the
// compute backend and a container is opened from it.
func (r *Builder) importScratch(ctx context.Context, opts compute.RuntimeOptions) (*compute.Container, error) {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(writeScratchTar(pw))
	}()
	ref, err := r.sess.Import(ctx, pr)
	if err != nil {
		return nil, err
	}
	return r.sess.Load(ctx, ref, opts)
}

// Compiles grants for this stage into an OCI runtime spec.
//
// A fresh [AffordanceBuilder] is created per stage so grant state does not
// bleed across stages. Reference grants are resolved and inlined recursively;
// domain grants are dispatched to the matching subsystem.
func (r *Builder) applyGrants(ctx context.Context, scopes []manifest.GrantScope) (*specs.Spec, error) {
	b := NewAffordanceBuilder()
	for _, scope := range scopes {
		if scope.Platform != "" && !matchesBuildPlatform(scope.Platform) {
			continue
		}
		for _, g := range scope.Grants {
			if err := b.Build(ctx, g, r.src); err != nil {
				return nil, err
			}
		}
	}
	return b.Spec().OCI, nil
}

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
		shell = "/bin/sh"
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
		Stdout:  newStreamWriter(ctx, slog.LevelInfo, "stdout"),
		Stderr:  newStreamWriter(ctx, slog.LevelInfo, "stderr"),
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
		return crex.Wrapf(ErrBuild, "malformed copy step %q: expected \"src dest\"", step.Copy)
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
		pw.CloseWithError(writeCopyTar(pw, filepath.Join(b.workdir, src), dest))
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
		return crex.Wrapf(ErrBuild, "unknown stage %q in cross-stage copy", stageName)
	}

	if !path.IsAbs(srcPath) {
		cfg, err := b.sess.Inspect(ctx, stageRef)
		if err != nil {
			return crex.Wrapf(ErrBuild, "get config of stage %q: %w", stageName, err)
		}
		workdir := cfg.WorkingDir
		if workdir == "" {
			workdir = "/"
		}
		srcPath = path.Join(workdir, srcPath)
	}

	rc, err := b.sess.Extract(ctx, stageRef, srcPath)
	if err != nil {
		return crex.Wrapf(ErrBuild, "extract %q from stage %q: %w", srcPath, stageName, err)
	}

	pr, pw := io.Pipe()
	go func() {
		err := rewriteTarPaths(pw, rc, srcPath, destPath)
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
		return crex.Wrapf(ErrBuild, "read config: %w", err)
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
