package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
	es "github.com/evanw/esbuild/pkg/api"
)

// [Builder] for Crucible widgets.
//
// Widgets are client-side JavaScript bundles built with esbuild.
type WidgetBuilder struct {
	registry         string
	defaultNamespace string
}

// Returns a [WidgetBuilder] configured with the given registry and namespace
// fallbacks for push operations.
func NewWidgetBuilder(registry, defaultNamespace string) *WidgetBuilder {
	return &WidgetBuilder{
		registry:         registry,
		defaultNamespace: defaultNamespace,
	}
}

// Builds a Crucible widget based on the provided manifest.
//
// It converts the manifest options into esbuild build options, invokes
// esbuild to perform the build, and processes the build result to log
// messages and handle errors.
func (wb *WidgetBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	widget, ok := m.Config.(*manifest.Widget)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	esOptions, err := esBuildOptionsFromManifest(widget, output)
	if err != nil {
		return nil, err
	}

	// esbuild doesn't support context cancellation, so this is the last chance
	// to abort the build.
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	result := es.Build(esOptions)

	if err := processEsBuildResult(result); err != nil {
		return nil, err
	}

	if _, err := m.ResolveName(wb.registry, wb.defaultNamespace); err != nil {
		return nil, crex.UserError("invalid resource name", "could not resolve the resource identifier").
			Fallback("Check the resource name in crucible.yaml.").
			Cause(err).
			Err()
	}

	if err := WriteManifest(&m, output); err != nil {
		return nil, err
	}

	return &BuildResult{
		Output:   output,
		Manifest: &m,
	}, nil
}

// Converts [manifest.Widget] options into esbuild's [es.BuildOptions].
//
// It maps the relevant fields and sets appropriate defaults for building
// Crucible widgets. The defaults are chosen to optimize for typical Crucible
// use cases, such as JSX support. Logging is disabled as we handle it ourselves.
// The Crucible UI library is marked as external to avoid bundling it, and the
// JSX factory and fragment are set to use Crucible's implementations.
//
// The project can include JavaScript (.js/.jsx) and/or TypeScript (.ts/.tsx)
// files. esbuild performs no type checking, even when TypeScript is used. To
// enforce type safety, run `tsc` separately before invoking esbuild.
//
// File syntax is inferred from extensions. If a tsconfig.json is present,
// esbuild respects only a subset of its options: "extends" (for configuration
// inheritance) and the "module" and "target" properties under "compilerOptions"
// (to set the output module format and JavaScript version, respectively). JSX
// options in tsconfig.json are not respected, as they are overridden to use
// Crucible’s custom JSX factory and fragment.
//
// For output, although esbuild supports CommonJS, ESM, and IIFE/UMD formats,
// Crucible supports only ESM output. Other formats are unlikely to be added.
// The build emits ES2015-compatible code to maintain broad environment support.
//
// Currently, crux builds only for web platforms. If additional platforms are
// introduced, the build process must run separately for each platform target.
func esBuildOptionsFromManifest(options *manifest.Widget, dist string) (es.BuildOptions, error) {

	// Determine project root
	projectRoot, err := filepath.Abs(filepath.Dir(options.Main))
	if err != nil {
		return es.BuildOptions{}, crex.Wrap(ErrInvalidPath, err)
	}

	esOptions := es.BuildOptions{

		// We handle logging ourselves
		LogLevel: es.LogLevelSilent,

		// Input
		AbsWorkingDir:     projectRoot,
		EntryPoints:       []string{options.Main},
		ResolveExtensions: []string{".tsx", ".ts", ".jsx", ".js"},
		Loader: map[string]es.Loader{
			".js":   es.LoaderJS,
			".jsx":  es.LoaderJSX,
			".ts":   es.LoaderTS,
			".tsx":  es.LoaderTSX,
			".yml":  es.LoaderNone,
			".yaml": es.LoaderNone,
		},

		// Output
		External: []string{
			"@cruciblehq/ui",
			"@cruciblehq/ui-web",
			"react",
			"react-reconciler",
		},
		Outdir:    dist,
		Platform:  es.PlatformBrowser,
		Target:    es.ES2015,
		Format:    es.FormatESModule,
		Sourcemap: es.SourceMapNone,
		Bundle:    true,
		Metafile:  true,
		Write:     true,
		Banner: map[string]string{
			"js": `import { __Crucible_createElement } from "@cruciblehq/ui";`,
		},

		// Optimizations
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
		TreeShaking:       es.TreeShakingTrue,

		// JSX
		JSX:         es.JSXTransform,
		JSXFactory:  "__Crucible_createElement",
		JSXFragment: "__Crucible_createElement",

		// Plugins
		Plugins: []es.Plugin{
			esplugin,
		},
	}

	return esOptions, nil
}

// Validates that the build directory contains the expected widget artifacts.
//
// A valid widget build directory must contain index.js.
func (wb *WidgetBuilder) Validate(buildDir string) error {
	manifestPath := filepath.Join(buildDir, manifest.ManifestFile)
	if _, err := os.Stat(manifestPath); err != nil {
		return crex.UserError("manifest not found", "build/crucible.yaml does not exist").
			Fallback("Run 'crux build' first to generate the build artifacts.").
			Cause(err).
			Err()
	}

	mainPath := filepath.Join(buildDir, manifest.WidgetMainFile)
	if _, err := os.Stat(mainPath); err != nil {
		return crex.UserError("widget build output not found", "build/index.js does not exist").
			Fallback("Run 'crux build' to generate the widget bundle.").
			Cause(err).
			Err()
	}

	return nil
}

// Packages the widget's build output into a distributable archive.
//
// The build directory must contain index.js.
func (wb *WidgetBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	if err := wb.Validate(buildDir); err != nil {
		return nil, err
	}
	return pack(ctx, buildDir, output)
}

// Uploads a widget package archive to the Hub registry.
//
// packagePath must point to an archive created by [WidgetBuilder.Pack].
func (wb *WidgetBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, m, packagePath, wb.registry, wb.defaultNamespace)
}
