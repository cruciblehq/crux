package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Common state and build logic for resource types that embed recipes.
type recipeBuilder struct {
	source  Source // Default registry and namespace for resolving references.
	workdir string // Directory containing the manifest, root for resolving copy sources.
}

// Builds a recipe by resolving sources and executing it against the runtime.
func (b *recipeBuilder) build(ctx context.Context, m manifest.Manifest, recipe *manifest.Recipe, output string, entrypoint []string) (*BuildResult, error) {
	return nil, crex.ProgrammingError("build failed", "recipe execution not implemented").
		Fallback("Recipe builds must be rewritten to execute remotely against the VM runtime.").
		Err()
}
