package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/reference"
)

// [Builder] for Crucible services.
//
// Embeds [recipeBuilder] for the build pipeline.
type ServiceBuilder struct {
	recipeBuilder
}

// Returns a [ServiceBuilder].
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewServiceBuilder(source Source, workdir string) *ServiceBuilder {
	return &ServiceBuilder{
		recipeBuilder: recipeBuilder{
			source:  source,
			workdir: workdir,
		},
	}
}

// Builds a Crucible service resource based on the provided manifest.
//
// The service configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (sb *ServiceBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifestConfig[*manifest.Service](&m)
	if err != nil {
		return nil, err
	}

	if _, err := reference.ParseIdentifier(m.Resource.Name, string(m.Resource.Type)); err != nil {
		return nil, crex.UserError("invalid resource name", "could not parse the resource identifier").
			Fallback("Check the resource name in crucible.yaml.").
			Cause(err).
			Err()
	}

	result, err := sb.build(ctx, m, &cfg.Recipe, output, cfg.Entrypoint)
	if err != nil {
		return nil, err
	}

	if err := WriteManifest(&m, result.Output); err != nil {
		return nil, err
	}

	return result, nil
}

// Verifies that the build directory contains the expected service artifacts.
func (sb *ServiceBuilder) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeService, manifest.ImageFile)
}

// Packages the service's build output into a distributable archive.
func (sb *ServiceBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads a service package archive to the Hub registry.
func (sb *ServiceBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, sb.source.Registry, m, packagePath)
}
