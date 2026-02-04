package build

import (
	"context"

	"github.com/cruciblehq/protocol/pkg/crex"
	"github.com/cruciblehq/protocol/pkg/manifest"
)

// Builder for Crucible runtimes.
type RuntimeBuilder struct {
	registry string // Hub registry URL for pulling base runtimes.
}

// Creates a new instance of [RuntimeBuilder].
func NewRuntimeBuilder(registry string) *RuntimeBuilder {
	return &RuntimeBuilder{registry: registry}
}

// Builds a Crucible runtime resource based on the provided manifest.
//
// Runtime resources are base images containing interpreters or other execution
// environments. If a base is specified, extends it; otherwise builds from
// scratch requiring platform-specific files. Outputs to the standardized
// dist/ location.
func (rb *RuntimeBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*Result, error) {
	runtime, ok := m.Config.(*manifest.Runtime)
	if !ok {
		return nil, crex.ProgrammingError("an internal configuration type mismatch occurred", "unexpected manifest type").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	if err := rb.validateManifest(runtime); err != nil {
		return nil, err
	}

	return NewImageBuilder(rb.registry, runtime.Base, runtime.Files, nil, output).Build(ctx)
}

// Validates required fields in the runtime manifest.
func (rb *RuntimeBuilder) validateManifest(_ *manifest.Runtime) error {
	return nil // No required fields for runtimes at this time.
}
