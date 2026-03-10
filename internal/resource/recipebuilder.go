package resource

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/protocol"
)

// Common state and build logic for resource types that embed recipes.
type recipeBuilder struct {
	client  BuildClient // Daemon client for sending build requests.
	source  Source      // Default registry and namespace for references.
	workdir string      // Directory containing the manifest, root for resolving copy sources.
}

// Builds a recipe by resolving sources and delegating execution to cruxd.
//
// All stage sources are resolved to local file paths (pulling and extracting
// remote references as needed). The resolved recipe is sent to the daemon
// as a [protocol.BuildRequest]. The daemon handles container creation, step
// execution, and image export.
func (b *recipeBuilder) build(ctx context.Context, m manifest.Manifest, recipe *manifest.Recipe, output string, entrypoint []string) (*BuildResult, error) {
	if b.client == nil {
		return nil, crex.ProgrammingError("build failed", "daemon client is nil").
			Fallback("Ensure a compute node is running and pass a daemon client when constructing this builder.").
			Err()
	}

	resolved, err := resolveAllSources(ctx, recipe, b.source)
	if err != nil {
		return nil, err
	}

	req := &protocol.BuildRequest{
		Recipe:     resolved,
		Resource:   m.Resource.Name,
		Output:     output,
		Root:       b.workdir,
		Entrypoint: entrypoint,
	}

	result, err := b.client.Build(ctx, req)
	if err != nil {
		return nil, crex.Wrap(ErrBuild, err)
	}

	return &BuildResult{Output: result.Output, Manifest: &m}, nil
}

// Resolves all stage sources in a recipe to forms the daemon can handle.
//
// Crucible runtime references (space-separated name and version constraint)
// are pulled, extracted, and rewritten to file paths. File and OCI sources
// are passed through unchanged — OCI references (single-token image names)
// are resolved by the daemon at build time. Returns a copy of the recipe
// with resolved sources and any error encountered during resolution.
func resolveAllSources(ctx context.Context, recipe *manifest.Recipe, source Source) (*manifest.Recipe, error) {
	resolved := *recipe
	resolved.Stages = make([]manifest.Stage, len(recipe.Stages))
	copy(resolved.Stages, recipe.Stages)

	for i := range resolved.Stages {
		stage := &resolved.Stages[i]

		src, err := stage.ParseFrom()
		if err != nil {
			return nil, crex.Wrapf(ErrBuild, "stage %s: %w", stageLabel(stage.Name, i), err)
		}

		if src.Type == manifest.SourceRef {
			filePath, _, err := source.Resolve(ctx, manifest.TypeRuntime, src.Value)
			if err != nil {
				return nil, crex.Wrapf(ErrBuild, "stage %s: %w", stageLabel(stage.Name, i), err)
			}
			stage.From = fmt.Sprintf("file %s", filePath)
		}
	}

	return &resolved, nil
}

// Returns a label for a stage, preferring the name when available and falling
// back to the 1-based index.
func stageLabel(name string, index int) string {
	if name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fmt.Sprintf("%d", index+1)
}
