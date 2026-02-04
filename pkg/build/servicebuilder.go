package build

import (
	"context"

	"github.com/cruciblehq/protocol/pkg/crex"
	"github.com/cruciblehq/protocol/pkg/manifest"
)

// Builder for Crucible services.
type ServiceBuilder struct {
	registry string // Hub registry URL for pulling runtimes.
}

// Creates a new instance of [ServiceBuilder].
func NewServiceBuilder(registry string) *ServiceBuilder {
	return &ServiceBuilder{registry: registry}
}

// Builds a Crucible service resource based on the provided manifest.
//
// Service resources extend a runtime base image with application code. This
// method resolves the runtime reference, loads the base image, adds service
// files as new layers, sets the entrypoint, and outputs a multi-platform OCI
// image to the standardized dist/ location.
func (sb *ServiceBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*Result, error) {
	service, ok := m.Config.(*manifest.Service)
	if !ok {
		return nil, crex.ProgrammingError("an internal configuration type mismatch occurred", "unexpected manifest type").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	if err := sb.validateManifest(service); err != nil {
		return nil, err
	}

	return NewImageBuilder(sb.registry, service.Runtime, service.Files, service.Entrypoint, output).Build(ctx)
}

// Validates required fields in the service manifest, requiring a runtime
// reference and an entrypoint.
func (sb *ServiceBuilder) validateManifest(service *manifest.Service) error {
	if service.Runtime == "" {
		return crex.UserError("runtime not specified", "service manifest has no runtime").
			Fallback("Add a runtime reference to the service manifest.").
			Err()
	}
	if len(service.Entrypoint) == 0 {
		return crex.UserError("entrypoint not specified", "service manifest has no entrypoint").
			Fallback("Add an entrypoint to the service manifest specifying the command to run.").
			Err()
	}
	return nil
}
