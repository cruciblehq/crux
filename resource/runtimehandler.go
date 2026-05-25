package resource

import (
	"context"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/resource/recipe"
	"github.com/cruciblehq/crux/source"
)

// Handler for Crucible runtimes.
//
// Extracts the runtime configuration from the manifest and delegates to the
// shared recipe pipeline. Only the Build and Pack operations are supported.
type RuntimeHandler struct {
	src     source.Source // Provides registry access for resolving references during builds.
	workdir string        // Manifest directory; root for resolving copy step sources.
}

// Returns a [RuntimeHandler].
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewRuntimeHandler(src source.Source, workdir string) *RuntimeHandler {
	return &RuntimeHandler{
		src:     src,
		workdir: workdir,
	}
}

// Builds a Crucible runtime resource based on the provided manifest.
//
// The runtime configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (rh *RuntimeHandler) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Runtime](&m)
	if err != nil {
		return nil, err
	}

	backend, err := compute.NewLocalImageBuilder()
	if err != nil {
		return nil, err
	}
	defer backend.Close()

	builder := recipe.NewBuilder(rh.src, rh.workdir, backend)
	buildDir, err := builder.Run(ctx, m, &cfg.Recipe, output, nil)
	if err != nil {
		return nil, err
	}

	if err := manifest.WriteAt(&m, buildDir); err != nil {
		return nil, err
	}

	return &BuildResult{Output: buildDir, Manifest: &m}, nil
}

// Verifies that the build directory contains the expected runtime artifacts.
func (rh *RuntimeHandler) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeRuntime, paths.ImageFile)
}

// Packages the runtime's build output into a distributable archive.
func (rh *RuntimeHandler) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads a runtime package archive to the Hub registry.
func (rh *RuntimeHandler) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return rh.src.Push(ctx, m.Resource.Name, string(m.Resource.Type), m.Resource.Version, packagePath)
}
