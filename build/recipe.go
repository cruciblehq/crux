package build

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/kit/archive"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/pack"
	"github.com/cruciblehq/crux/pull"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Holds shared state for building all stages of a recipe.
type recipeBuild struct {
	client  *containerd.Client          // Shared containerd gRPC connection.
	id      *reference.Identifier       // Parsed resource identifier.
	version string                      // Resource version from the manifest, used as the base for stage image versions.
	options reference.IdentifierOptions // Options for parsing references in the recipe.
	output  string                      // Output directory for the final build artifact.
}

// Builds a recipe end-to-end against the container runtime.
//
// This is the shared pipeline for all resource types that embed a recipe.
// All stages are built in declaration order. The non-transient stage is
// exported as the final image.
func buildRecipe(ctx context.Context, m manifest.Manifest, recipe *manifest.Recipe, registry, defaultNamespace, output string) (*Result, error) {
	options := reference.IdentifierOptions{
		DefaultRegistry:  registry,
		DefaultNamespace: defaultNamespace,
	}

	id, err := reference.ParseIdentifier(m.Resource.Name, m.Resource.Type, options)
	if err != nil {
		return nil, err
	}

	client, err := runtime.NewContainerdClient(id.Hostname())
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, done, err := client.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)

	rb := &recipeBuild{
		client:  client,
		id:      id,
		version: m.Resource.Version,
		options: options,
		output:  output,
	}

	for i, stage := range recipe.Stages {
		if err := rb.buildStage(ctx, stage, i); err != nil {
			return nil, fmt.Errorf("stage %s: %w", stageLabel(stage.Name, i), err)
		}
	}

	return &Result{Output: output, Manifest: &m}, nil
}

// Builds a single stage of a recipe.
//
// Resolves the stage's base image, starts a build container, executes the
// stage's steps, then commits the result. Non-transient stages are exported
// to the output directory.
func (rb *recipeBuild) buildStage(ctx context.Context, stage manifest.Stage, index int) error {
	label := stageLabel(stage.Name, index)
	slog.Info("building stage", "stage", label)

	source, err := parseFrom(stage.From, rb.options)
	if err != nil {
		return err
	}

	version := rb.version
	if !stage.Transient {
		version = stageVersion(version, label)
	}
	img := runtime.NewImage(rb.client, rb.id, version)

	if err := rb.resolveSource(ctx, img, source); err != nil {
		return err
	}

	ctr, err := img.Start(ctx, "")
	if err != nil {
		return err
	}
	defer ctr.Destroy(ctx)

	if err := executeSteps(ctx, ctr, stage.Steps, newRecipeState()); err != nil {
		return err
	}

	if err := ctr.Stop(ctx); err != nil {
		return err
	}

	if err := ctr.Commit(ctx); err != nil {
		return err
	}

	if !stage.Transient {
		return img.Export(ctx, filepath.Join(rb.output, pack.ImageFile))
	}

	return nil
}

// Resolves a recipe source into an imported container image.
//
// For file sources the local OCI tarball is imported directly. For ref
// sources the runtime archive is pulled from the registry (with caching),
// extracted to a temporary directory, and the contained image.tar is
// imported.
func (rb *recipeBuild) resolveSource(ctx context.Context, img *runtime.Image, src source) error {
	switch src.Type {
	case sourceFile:
		return img.Import(ctx, src.Value)
	case sourceRef:
		return rb.resolveRefSource(ctx, img, src)
	}
	return nil
}

// Pulls a runtime archive from the registry, extracts image.tar, and imports
// it into the container runtime.
func (rb *recipeBuild) resolveRefSource(ctx context.Context, img *runtime.Image, src source) error {
	result, err := pull.Pull(ctx, pull.Options{
		Registry:         rb.options.DefaultRegistry,
		Reference:        src.Value,
		Type:             resource.TypeRuntime,
		DefaultNamespace: rb.options.DefaultNamespace,
	})
	if err != nil {
		return err
	}

	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return err
	}
	defer localCache.Close()

	archiveReader, err := localCache.OpenArchive(ctx, result.Namespace, result.Resource, result.Version)
	if err != nil {
		return err
	}
	defer archiveReader.Close()

	extractDir, err := os.MkdirTemp("", "crux-runtime-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	if err := archive.ExtractFromReader(archiveReader, extractDir, archive.Zstd); err != nil {
		return err
	}

	imagePath := filepath.Join(extractDir, pack.ImageFile)
	return img.Import(ctx, imagePath)
}

// Returns a version string for a transient stage image, distinguishing it
// from the output image.
func stageVersion(version, label string) string {
	return fmt.Sprintf("%s-stage-%s", version, label)
}

// Returns a human-readable label for a stage, preferring the name when
// available and falling back to the 1-based index.
func stageLabel(name string, index int) string {
	if name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fmt.Sprintf("%d", index+1)
}

// Executes a list of steps in order against the build container.
func executeSteps(ctx context.Context, ctr *runtime.Container, steps []manifest.Step, state *recipeState) error {
	for i, step := range steps {
		if err := executeStep(ctx, ctr, step, state); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return nil
}

// Executes a single step, dispatching to operation execution, group recursion,
// or state mutation depending on the step's fields.
func executeStep(ctx context.Context, ctr *runtime.Container, step manifest.Step, state *recipeState) error {
	hasOp := step.Run != "" || step.Copy != ""

	// Platform group: apply group-level modifiers and recurse.
	if len(step.Steps) > 0 {
		state.apply(step)
		return executeSteps(ctx, ctr, step.Steps, state)
	}

	// Operation with optional scoped modifiers.
	if hasOp {
		return executeOperation(ctx, ctr, step, state)
	}

	// Standalone modifier(s): persist in state.
	state.apply(step)
	return nil
}

// Executes a run or copy operation with scoped modifier overrides.
//
// Step-level modifiers override the persistent state for this operation only.
// The persistent state is not modified.
func executeOperation(ctx context.Context, ctr *runtime.Container, step manifest.Step, state *recipeState) error {
	shell, opts := state.resolve(step)

	if err := ensureDir(ctx, ctr, opts.Workdir); err != nil {
		return err
	}

	switch {
	case step.Run != "":
		slog.Debug("run", "command", step.Run, "shell", shell)
		result, err := ctr.ExecWith(ctx, opts, shell, "-c", step.Run)
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("command failed (exit %d): %s", result.ExitCode, result.Stderr)
		}

	case step.Copy != "":
		// TODO: copy file into container filesystem.
		slog.Warn("copy not yet implemented", "copy", step.Copy)
	}

	return nil
}

// Creates a directory inside the container. No-op if dir is empty.
func ensureDir(ctx context.Context, ctr *runtime.Container, dir string) error {
	if dir == "" {
		return nil
	}
	result, err := ctr.Exec(ctx, "mkdir", "-p", dir)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to create workdir %q (exit %d): %s", dir, result.ExitCode, result.Stderr)
	}
	return nil
}
