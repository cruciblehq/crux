package recipe

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/resource/affordance"
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

// Runs a single stage: imports the base image, compiles affordances into a
// security policy, and executes each step in order.
//
// stageImages is updated with the final image ref under the stage name when
// the stage has a name, making it available as a copy source for later stages.
func (r *Builder) runStage(ctx context.Context, num int, stage *manifest.Stage, stageImages map[string]string) (string, error) {
	currentRef, err := r.importBase(ctx, stage)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "stage %d: import base: %w", num, err)
	}

	var security *specs.Spec
	if len(stage.Grants) > 0 {
		security, err = r.applyGrants(ctx, stage.Grants)
		if err != nil {
			return "", crex.Wrapf(ErrBuild, "stage %d: compile grants: %w", num, err)
		}
	}

	state := &stageState{}
	for j := range stage.Steps {
		currentRef, err = r.executeStep(ctx, currentRef, &stage.Steps[j], security, state, stageImages)
		if err != nil {
			return "", crex.Wrapf(ErrBuild, "stage %d step %d: %w", num, j+1, err)
		}
	}

	if stage.Name != "" {
		stageImages[stage.Name] = currentRef
	}

	return currentRef, nil
}

// Resolves and imports the base image for a stage.
//
// When stage.From is nil the stage starts from scratch and a minimal empty
// OCI image is imported. When From is set, the referenced runtime is pulled
// and its image.tar is imported into the compute backend.
func (r *Builder) importBase(ctx context.Context, stage *manifest.Stage) (string, error) {
	if stage.From == "" {
		return r.importScratch(ctx)
	}

	ref, err := r.src.Parse(string(manifest.TypeRuntime), stage.From)
	if err != nil {
		return "", err
	}

	result, err := r.src.Pull(ctx, ref)
	if err != nil {
		return "", err
	}

	imgPath := filepath.Join(result.Extracted, paths.ImageFile)
	f, err := os.Open(imgPath)
	if err != nil {
		return "", crex.Wrap(ErrFileSystemOperation, err)
	}
	defer f.Close()

	return r.backend.Import(ctx, f)
}

// Creates and imports a minimal empty (scratch) OCI image.
//
// The image has no layers and an empty configuration. It is imported into the
// compute backend and its reference is returned.
func (r *Builder) importScratch(ctx context.Context) (string, error) {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(writeScratchTar(pw))
	}()
	return r.backend.Import(ctx, pr)
}

// Compiles grants for this stage into an OCI runtime spec.
//
// A fresh [affordance.Builder] is created per stage so grant state does not
// bleed across stages. Reference grants are resolved and inlined recursively;
// domain grants are dispatched to the matching subsystem.
func (r *Builder) applyGrants(ctx context.Context, scopes []manifest.GrantScope) (*specs.Spec, error) {
	b := affordance.NewBuilder()
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
// Run steps spawn a build container and commit the diff. Copy steps append a
// tar layer constructed from host files. Standalone modifier steps update the
// persistent stage state and optionally update the image configuration.
// Platform groups recurse over their children when the target platform matches.
func (b *Builder) executeStep(ctx context.Context, imageRef string, step *manifest.Step, security *specs.Spec, state *stageState, stageImages map[string]string) (string, error) {
	if step.Platform != "" && len(step.Steps) > 0 {
		return b.executePlatformGroup(ctx, imageRef, step, security, state, stageImages)
	}

	if step.Platform != "" && !matchesBuildPlatform(step.Platform) {
		return imageRef, nil
	}

	switch {
	case step.Run != "":
		return b.runStep(ctx, imageRef, step, security, state)
	case step.Copy != "":
		return b.copyStep(ctx, imageRef, step, state, stageImages)
	default:
		return b.applyModifierStep(ctx, imageRef, step, state)
	}
}

// Executes a platform-scoped group of steps when the platform matches.
//
// Group-level modifier fields are applied to the persistent stage state before
// the children are executed. If the platform does not match the current build
// target, all children are skipped.
func (b *Builder) executePlatformGroup(ctx context.Context, imageRef string, step *manifest.Step, security *specs.Spec, state *stageState, stageImages map[string]string) (string, error) {
	if !matchesBuildPlatform(step.Platform) {
		return imageRef, nil
	}
	if hasModifiers(step) {
		applyStateModifiers(state, step)
	}
	var err error
	for i := range step.Steps {
		imageRef, err = b.executeStep(ctx, imageRef, &step.Steps[i], security, state, stageImages)
		if err != nil {
			return "", err
		}
	}
	return imageRef, nil
}

// Executes a Run step inside a build container.
//
// The command is run with the effective shell, environment, workdir, and user
// from the persistent stage state merged with any step-local overrides. The
// filesystem diff is committed as a new image layer.
func (b *Builder) runStep(ctx context.Context, imageRef string, step *manifest.Step, security *specs.Spec, state *stageState) (string, error) {
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

	env := envMapToSlice(state.env)
	for k, v := range step.Env {
		env = append(env, k+"="+v)
	}

	cfg := &compute.RunConfig{
		Shell:    shell,
		Command:  step.Run,
		Env:      env,
		WorkDir:  workdir,
		User:     user,
		Security: security,
		Stdout:   newStreamWriter(ctx, slog.LevelInfo, "stdout"),
		Stderr:   newStreamWriter(ctx, slog.LevelInfo, "stderr"),
	}

	return b.backend.Run(ctx, imageRef, cfg)
}

// Executes a Copy step by appending a tar layer to the image.
//
// For local copies the host source path is resolved relative to the manifest
// directory. For cross-stage copies (src of the form "stageName:/path") the
// source is extracted from the named stage's image and rewritten to the
// destination path. The destination defaults to the current workdir when set,
// otherwise to the image root.
func (b *Builder) copyStep(ctx context.Context, imageRef string, step *manifest.Step, state *stageState, stageImages map[string]string) (string, error) {
	src, dest, ok := strings.Cut(step.Copy, " ")
	if !ok {
		return "", crex.Wrapf(ErrBuild, "malformed copy step %q: expected \"src dest\"", step.Copy)
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
		return b.crossStageCopy(ctx, imageRef, stageName, srcPath, dest, stageImages)
	}

	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(writeCopyTar(pw, filepath.Join(b.workdir, src), dest))
	}()

	return b.backend.Copy(ctx, imageRef, pr)
}

// Copies srcPath from the named prior stage into the current image at destPath.
//
// Extracts srcPath from the stage's final image, rewrites the tar paths from
// srcPath to destPath, and appends the result as a new layer.
func (b *Builder) crossStageCopy(ctx context.Context, imageRef, stageName, srcPath, destPath string, stageImages map[string]string) (string, error) {
	stageRef, ok := stageImages[stageName]
	if !ok {
		return "", crex.Wrapf(ErrBuild, "unknown stage %q in cross-stage copy", stageName)
	}

	rc, err := b.backend.Extract(ctx, stageRef, srcPath)
	if err != nil {
		return "", crex.Wrapf(ErrBuild, "extract %q from stage %q: %w", srcPath, stageName, err)
	}

	pr, pw := io.Pipe()
	go func() {
		err := rewriteTarPaths(pw, rc, srcPath, destPath)
		rc.Close()
		pw.CloseWithError(err)
	}()

	return b.backend.Copy(ctx, imageRef, pr)
}

// Applies a standalone modifier step.
//
// Updates the persistent stage state and, for fields visible in the final
// image (Env, Workdir, User), updates the image configuration as well. Shell
// changes are build-time only and not reflected in the image config.
func (b *Builder) applyModifierStep(ctx context.Context, imageRef string, step *manifest.Step, state *stageState) (string, error) {
	applyStateModifiers(state, step)

	update := &compute.ConfigUpdate{}
	if step.Workdir != "" {
		update.SetWorkDir = step.Workdir
	}
	if step.User != "" {
		update.SetUser = step.User
	}
	if len(step.Env) > 0 {
		update.AddEnv = step.Env
	}

	if update.SetWorkDir == "" && update.SetUser == "" && len(update.AddEnv) == 0 {
		return imageRef, nil
	}

	return b.backend.Configure(ctx, imageRef, update)
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

// Converts an env map to a "KEY=VALUE" slice.
func envMapToSlice(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
