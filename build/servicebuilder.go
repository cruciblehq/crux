package build

import (
	"context"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/manifest"
)

// Builder for Crucible services.
type ServiceBuilder struct {
	registry         string // Hub registry URL for resolving references.
	defaultNamespace string // Default namespace for resolving references.
	context          string // Project directory, containing the manifest.
}

// Creates a new instance of [ServiceBuilder].
func NewServiceBuilder(registry, defaultNamespace, context string) *ServiceBuilder {
	return &ServiceBuilder{
		registry:         registry,
		defaultNamespace: defaultNamespace,
		context:          context,
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

	return buildRecipe(ctx, m, &cfg.Recipe, sb.registry, sb.defaultNamespace, output, sb.context)
}
