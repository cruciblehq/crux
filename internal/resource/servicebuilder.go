package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
)

// [Builder] for Crucible services.
//
// Embeds [recipeBuilder] for the build pipeline.
type ServiceBuilder struct {
	recipeBuilder
}

// Returns a [ServiceBuilder] wired to the given daemon client.
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewServiceBuilder(client BuildClient, source Source, workdir string) *ServiceBuilder {
	return &ServiceBuilder{
		recipeBuilder: recipeBuilder{
			client:  client,
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
	cfg, ok := m.Config.(*manifest.Service)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
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

// Validates that the build directory contains the expected service artifacts.
//
// A valid service build directory must contain image.tar.
func (sb *ServiceBuilder) Validate(buildDir string) error {
	manifestPath := filepath.Join(buildDir, manifest.ManifestFile)
	if _, err := os.Stat(manifestPath); err != nil {
		return crex.UserError("manifest not found", "build/crucible.yaml does not exist").
			Fallback("Run 'crux build' first to generate the build artifacts.").
			Cause(err).
			Err()
	}

	imagePath := filepath.Join(buildDir, manifest.ImageFile)
	if _, err := os.Stat(imagePath); err != nil {
		return crex.UserError("service build output not found", "build/image.tar does not exist").
			Fallback("Run 'crux build' to prepare the service image.").
			Cause(err).
			Err()
	}

	return nil
}

// Packages the service's build output into a distributable archive.
//
// The build directory must contain image.tar.
func (sb *ServiceBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	if err := sb.Validate(buildDir); err != nil {
		return nil, err
	}
	return pack(ctx, buildDir, output)
}

// Uploads a service package archive to the Hub registry.
//
// packagePath must point to an archive created by [ServiceBuilder.Pack].
func (sb *ServiceBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, m, packagePath, sb.source)
}
