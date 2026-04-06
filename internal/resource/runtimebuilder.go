package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/reference"
)

// Builder for Crucible runtimes.
//
// Extracts the runtime configuration from the manifest and delegates to the
// shared recipe pipeline. Only the Build and Pack operations are supported.
type RuntimeBuilder struct {
	recipeBuilder
}

// Returns a [RuntimeBuilder].
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewRuntimeBuilder(source Source, workdir string) *RuntimeBuilder {
	return &RuntimeBuilder{
		recipeBuilder: recipeBuilder{
			source:  source,
			workdir: workdir,
		},
	}
}

// Builds a Crucible runtime resource based on the provided manifest.
//
// The runtime configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (rb *RuntimeBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifestConfig[*manifest.Runtime](&m)
	if err != nil {
		return nil, err
	}

	if _, err := reference.ParseIdentifier(m.Resource.Name, string(m.Resource.Type)); err != nil {
		return nil, crex.UserError("invalid resource name", "could not parse the resource identifier").
			Fallback("Check the resource name in crucible.yaml.").
			Cause(err).
			Err()
	}

	result, err := rb.build(ctx, m, &cfg.Recipe, output, nil)
	if err != nil {
		return nil, err
	}

	if err := WriteManifest(&m, result.Output); err != nil {
		return nil, err
	}

	return result, nil
}

// Verifies that the build directory contains the expected runtime artifacts.
func (rb *RuntimeBuilder) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeRuntime, manifest.ImageFile)
}

// Packages the runtime's build output into a distributable archive.
func (rb *RuntimeBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads a runtime package archive to the Hub registry.
func (rb *RuntimeBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, rb.source.Registry, m, packagePath)
}
