package build

import (
	"path/filepath"

	"github.com/cruciblehq/crux/pkg/manifest"
	es "github.com/evanw/esbuild/pkg/api"
)

// Builds the widget specified in the given options.
//
// It converts the build options to esbuild options, and invokes esbuild to
// perform the build process.
func BuildWidget(options *manifest.Widget) error {

	// Convert to esbuild options
	esOptions, err := esBuildOptionsFromManifest(options)
	if err != nil {
		return err
	}

	// Build with esbuild
	result := es.Build(esOptions)

	// Process build result
	return processBuildResult(result)
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
func esBuildOptionsFromManifest(options *manifest.Widget) (es.BuildOptions, error) {

	esOptions := es.BuildOptions{

		// We handle logging ourselves
		LogLevel: es.LogLevelSilent,

		// Input
		EntryPoints:       []string{options.Build.Main},
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
		Outdir:    filepath.Dir(options.Build.Dist),
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
