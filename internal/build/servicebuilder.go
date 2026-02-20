package build

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/manifest"
)

// Builder for Crucible services.
type ServiceBuilder struct {
	recipeBuilder
}

// Creates a new instance of [ServiceBuilder].
func NewServiceBuilder(client *daemon.Client, registry, defaultNamespace, context string) *ServiceBuilder {
	return &ServiceBuilder{
		recipeBuilder: recipeBuilder{
			client:           client,
			registry:         registry,
			defaultNamespace: defaultNamespace,
			context:          context,
		},
	}
}

// Builds a Crucible service resource based on the provided manifest.
//
// The service configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (sb *ServiceBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*Result, error) {
	cfg, ok := m.Config.(*manifest.Service)
	if !ok {
		return nil, crex.ProgrammingError("an internal configuration type mismatch occurred", "unexpected manifest type").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	return sb.build(ctx, m, &cfg.Recipe, output, cfg.Entrypoint)
}
