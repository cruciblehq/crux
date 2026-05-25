// Package widget implements the esbuild-based build pipeline for widgets.
//
// [Build] compiles a widget source tree into a browser-ready ESM bundle using
// esbuild. Module resolution, JSX transformation, and build result processing
// are all handled here.
//
//	dist, err := widget.Build(ctx, cfg, "build/widget")
package widget
