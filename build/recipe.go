package build

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/runtime"
)

// Builds a recipe end-to-end against the container runtime.
//
// This is the shared pipeline for all resource types that embed a recipe.
// It resolves the base image from the recipe's from field, starts a build
// container, executes every step in order, and returns the result.
func buildRecipe(ctx context.Context, m manifest.Manifest, recipe *manifest.Recipe, registry, defaultNamespace, output string) (*Result, error) {
	options := reference.IdentifierOptions{
		DefaultRegistry:  registry,
		DefaultNamespace: defaultNamespace,
	}

	source, err := recipe.ParseFrom(options)
	if err != nil {
		return nil, err
	}

	img, err := resolveSource(ctx, m, source, options)
	if err != nil {
		return nil, err
	}

	ctr, err := img.Start(ctx, "")
	if err != nil {
		return nil, err
	}
	defer ctr.Destroy(ctx)

	if err := executeSteps(ctx, ctr, recipe.Steps, newRecipeState()); err != nil {
		return nil, err
	}

	// TODO: commit container snapshot and export as OCI tarball to output.

	return &Result{Output: output, Manifest: &m}, nil
}

// Resolves a recipe source into an imported container image.
//
// For file sources the local OCI tarball is imported directly. Ref sources
// are not yet implemented.
func resolveSource(ctx context.Context, m manifest.Manifest, source manifest.RuntimeSource, options reference.IdentifierOptions) (*runtime.Image, error) {
	id, err := reference.ParseIdentifier(m.Resource.Name, m.Resource.Type, options)
	if err != nil {
		return nil, err
	}

	img := runtime.NewImage(id, m.Resource.Version)

	switch source.Type {
	case manifest.RuntimeSourceFile:
		if err := img.Import(ctx, source.Value); err != nil {
			return nil, err
		}
	case manifest.RuntimeSourceRef:
		// TODO: pull archive via registry.Client.DownloadArchive (or pull.Pull
		// for caching), extract image.tar from the tar.zst archive, then call
		// img.Import with the extracted path.
		return nil, fmt.Errorf("ref source not yet implemented")
	}

	return img, nil
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
