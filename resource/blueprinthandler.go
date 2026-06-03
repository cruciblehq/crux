package resource

import (
	"context"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/source"
)

// Handler for Crucible blueprints.
//
// Building a blueprint resolves service references and their affordances,
// producing a deployment plan as the build artifact. The plan is written
// to the output directory alongside the resolved manifest.
type BlueprintHandler struct {
	src         source.Source // Provides registry access for pulling service and runtime refs.
	environment string        // Selects which blueprint environment to include in the plan.
}

// Returns a new [BlueprintHandler].
//
// The environment selects which blueprint environment to include in the plan.
// It must match one of the environment IDs declared in the blueprint.
func NewBlueprintHandler(src source.Source, environment string) *BlueprintHandler {
	return &BlueprintHandler{src: src, environment: environment}
}

// Builds a Crucible blueprint resource based on the provided manifest.
//
// References are pulled from the registry and their runtime affordances are
// resolved into primitives. The resulting plan is written to the output
// directory as plan.yaml alongside the resolved manifest.
func (bh *BlueprintHandler) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Blueprint](&m)
	if err != nil {
		return nil, err
	}

	if err := Build(ctx, cfg, bh.environment, bh.src, output); err != nil {
		return nil, err
	}

	if err := manifest.WriteAt(&m, output); err != nil {
		return nil, err
	}

	return &BuildResult{Output: output, Manifest: &m}, nil
}

// Verifies that the build directory contains the expected blueprint artifacts.
func (bh *BlueprintHandler) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeBlueprint, files.PlanFile)
}

// Packages the blueprint's build output into a distributable archive.
func (bh *BlueprintHandler) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads a blueprint package archive to the Hub registry.
func (bh *BlueprintHandler) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return bh.src.Push(ctx, m.Resource.Name, string(m.Resource.Type), m.Resource.Version, packagePath)
}
