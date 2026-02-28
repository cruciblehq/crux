package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/manifest"
)

// Builder for Crucible runtimes.
//
// Extracts the runtime configuration from the manifest and delegates to the
// shared recipe pipeline. Only the Build and Pack operations are supported.
type RuntimeBuilder struct {
	recipeBuilder
}

// Returns a [RuntimeBuilder] wired to the given daemon client.
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewRuntimeBuilder(client *daemon.Client, registry, defaultNamespace, workdir string) *RuntimeBuilder {
	return &RuntimeBuilder{
		recipeBuilder: recipeBuilder{
			client:           client,
			registry:         registry,
			defaultNamespace: defaultNamespace,
			workdir:          workdir,
		},
	}
}

// Builds a Crucible runtime resource based on the provided manifest.
//
// The runtime configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (rb *RuntimeBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, ok := m.Config.(*manifest.Runtime)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	if _, err := m.ResolveName(rb.registry, rb.defaultNamespace); err != nil {
		return nil, crex.UserError("invalid resource name", "could not resolve the resource identifier").
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

// Validates that the build directory contains the expected runtime artifacts.
//
// A valid runtime build directory must contain image.tar.
func (rb *RuntimeBuilder) Validate(buildDir string) error {
	manifestPath := filepath.Join(buildDir, manifest.ManifestFile)
	if _, err := os.Stat(manifestPath); err != nil {
		return crex.UserError("manifest not found", "build/crucible.yaml does not exist").
			Fallback("Run 'crux build' first to generate the build artifacts.").
			Cause(err).
			Err()
	}

	imagePath := filepath.Join(buildDir, manifest.ImageFile)
	if _, err := os.Stat(imagePath); err != nil {
		return crex.UserError("runtime build output not found", "build/image.tar does not exist").
			Fallback("Run 'crux build' to prepare the runtime image.").
			Cause(err).
			Err()
	}

	return nil
}

// Packages the runtime's build output into a distributable archive.
//
// The build directory must contain image.tar.
func (rb *RuntimeBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	if err := rb.Validate(buildDir); err != nil {
		return nil, err
	}
	return pack(ctx, buildDir, output)
}

// Uploads a runtime package archive to the Hub registry.
//
// packagePath must point to an archive created by [RuntimeBuilder.Pack].
func (rb *RuntimeBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, m, packagePath, rb.registry, rb.defaultNamespace)
}
