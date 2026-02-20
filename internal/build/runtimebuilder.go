package build

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/manifest"
)

// Builder for Crucible runtimes.
//
// Extracts the runtime configuration from the manifest and delegates to the
// shared recipe pipeline.
type RuntimeBuilder struct {
	recipeBuilder
}

// Creates a new instance of [RuntimeBuilder].
func NewRuntimeBuilder(client *daemon.Client, registry, defaultNamespace, context string) *RuntimeBuilder {
	return &RuntimeBuilder{
		recipeBuilder: recipeBuilder{
			client:           client,
			registry:         registry,
			defaultNamespace: defaultNamespace,
			context:          context,
		},
	}
}

// Builds a Crucible runtime resource based on the provided manifest.
//
// The runtime configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (rb *RuntimeBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*Result, error) {
	cfg, ok := m.Config.(*manifest.Runtime)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	return rb.build(ctx, m, &cfg.Recipe, output, nil)
}
