package resource

import (
	"context"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/source"
)

// Handler for affordances.
//
// Building an affordance resolves its grant list. References (grants with no
// Subsystem) are pulled from the registry and inlined. Domain grants are
// dispatched to the appropriate subsystem. The output is a resolved manifest
// containing only built domain grants and groups with no remaining references.
type AffordanceHandler struct {
	src source.Source // Provides registry access.
}

// Returns an [AffordanceHandler] configured with the given source.
func NewAffordanceHandler(src source.Source) *AffordanceHandler {
	return &AffordanceHandler{src: src}
}

// Builds an affordance based on the provided manifest.
//
// Affordance references in the input list are walked and expanded recursively.
// The resolved manifest written to the output directory contains only built
// domain grants with no further references.
func (ah *AffordanceHandler) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Affordance](&m)
	if err != nil {
		return nil, err
	}

	b := NewAffordanceBuilder()
	for _, scope := range cfg.Scopes {
		for _, g := range scope.Grants {
			if err := b.Build(ctx, g, ah.src); err != nil {
				return nil, err
			}
		}
	}

	m.Config = &manifest.Affordance{
		Schema: cfg.Schema,
		Scopes: cfg.Scopes,
	}

	if err := manifest.WriteAt(&m, output); err != nil {
		return nil, err
	}

	return &BuildResult{Output: output, Manifest: &m}, nil
}

// Verifies that the build directory contains the expected affordance artifacts.
func (ah *AffordanceHandler) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeAffordance, "")
}

// Packages the affordance's build output into a distributable archive.
func (ah *AffordanceHandler) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads an affordance package archive to the Hub registry.
func (ah *AffordanceHandler) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return ah.src.Push(ctx, m.Resource.Name, string(m.Resource.Type), m.Resource.Version, packagePath)
}
