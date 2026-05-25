package resource

import (
	"context"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/resource/recipe"
	"github.com/cruciblehq/crux/source"
)

// Handler for Crucible services.
//
// Extracts the service configuration from the manifest and delegates to the
// shared recipe pipeline.
type ServiceHandler struct {
	src     source.Source // Provides registry access for resolving references during builds.
	workdir string        // Manifest directory; root for resolving copy step sources.
}

// Returns a [ServiceHandler].
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewServiceHandler(src source.Source, workdir string) *ServiceHandler {
	return &ServiceHandler{
		src:     src,
		workdir: workdir,
	}
}

// Builds a Crucible service resource based on the provided manifest.
//
// The service configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (sh *ServiceHandler) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Service](&m)
	if err != nil {
		return nil, err
	}

	backend, err := compute.NewLocalImageBuilder()
	if err != nil {
		return nil, err
	}
	defer backend.Close()

	builder := recipe.NewBuilder(sh.src, sh.workdir, backend)
	buildDir, err := builder.Run(ctx, m, &cfg.Recipe, output, cfg.Entrypoint)
	if err != nil {
		return nil, err
	}

	if err := manifest.WriteAt(&m, buildDir); err != nil {
		return nil, err
	}

	return &BuildResult{Output: buildDir, Manifest: &m}, nil
}

// Verifies that the build directory contains the expected service artifacts.
func (sh *ServiceHandler) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeService, paths.ImageFile)
}

// Packages the service's build output into a distributable archive.
func (sh *ServiceHandler) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads a service package archive to the Hub registry.
func (sh *ServiceHandler) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return sh.src.Push(ctx, m.Resource.Name, string(m.Resource.Type), m.Resource.Version, packagePath)
}
