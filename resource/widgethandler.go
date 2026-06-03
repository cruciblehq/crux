package resource

import (
	"context"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/source"
)

// Handler for Crucible widgets.
//
// Widgets are client-side JavaScript bundles built with esbuild. The handler
// delegates to the widget package which converts the manifest configuration
// into esbuild options, invokes esbuild, and processes the result.
type WidgetHandler struct {
	src source.Source // Provides registry access for resolving references during builds.
}

// Returns a [WidgetHandler] configured with the given source for push operations.
func NewWidgetHandler(src source.Source) *WidgetHandler {
	return &WidgetHandler{src: src}
}

// Builds a Crucible widget based on the provided manifest.
//
// Delegates to [widget.Build] which converts the manifest options into esbuild
// build options, invokes esbuild, and processes the result.
func (wh *WidgetHandler) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifest.As[*manifest.Widget](&m)
	if err != nil {
		return nil, err
	}

	if _, err := BuildWidget(ctx, cfg, output); err != nil {
		return nil, err
	}

	if err := manifest.WriteAt(&m, output); err != nil {
		return nil, err
	}

	return &BuildResult{Output: output, Manifest: &m}, nil
}

// Verifies that the build directory contains the expected widget artifacts.
func (wh *WidgetHandler) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeWidget, files.WidgetMainFile)
}

// Packages the widget's build output into a distributable archive.
func (wh *WidgetHandler) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads a widget package archive to the Hub registry.
func (wh *WidgetHandler) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return wh.src.Push(ctx, m.Resource.Name, string(m.Resource.Type), m.Resource.Version, packagePath)
}
