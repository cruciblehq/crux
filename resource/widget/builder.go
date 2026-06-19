package widget

import (
	"context"
	"path/filepath"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
	es "github.com/evanw/esbuild/pkg/api"
)

// Builds a Crucible widget resource from its configuration.
//
// Bundling is performed by esbuild.
type Builder struct{}

// Returns a new widget [Builder].
func NewBuilder() *Builder {
	return &Builder{}
}

// Builds the widget described by cfg into the dist directory.
//
// Context cancellation is checked before invoking esbuild (which does not
// support cancellation itself). Returns the esbuild metafile JSON on success.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Widget, dist string) (string, error) {
	esOptions, err := buildOptions(cfg, dist)
	if err != nil {
		return "", err
	}

	// esbuild doesn't support context cancellation; last chance to abort.
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	result := es.Build(esOptions)

	if err := processResult(result); err != nil {
		return "", err
	}

	return result.Metafile, nil
}

// Converts [manifest.Widget] options into esbuild's [es.BuildOptions].
//
// Maps the relevant fields and sets appropriate defaults for building widgets.
// The defaults are chosen to optimize for typical Crucible use cases, such as
// JSX support. Logging is disabled as we handle it ourselves. The Crucible UI
// library is marked as external to avoid bundling it, and the JSX factory and
// fragment are set to use Crucible's implementations. The project can include
// JavaScript (.js/.jsx) and/or TypeScript (.ts/.tsx) paths. esbuild performs
// no type checking, even when TypeScript is used. To enforce type safety, tsc
// should be invoked separately. File syntax is inferred from extensions. If a
// tsconfig.json is present, esbuild respects only a subset of its options:
// "extends" (for configuration inheritance) and the "module" and "target"
// properties under "compilerOptions" (to set the output module format and
// JavaScript version, respectively). JSX options in tsconfig.json are not
// respected, as they are overridden to use Crucible's custom JSX factory and
// fragment. For output, although esbuild supports CommonJS, ESM, and IIFE/UMD
// formats, Crucible supports only ESM output. Other formats are unlikely to be
// added in the future. The build emits ES2015-compatible code to maintain broad
// environment support. Currently, crux builds only for web platforms.
func buildOptions(options *manifest.Widget, dist string) (es.BuildOptions, error) {
	projectRoot, err := filepath.Abs(filepath.Dir(options.Main))
	if err != nil {
		return es.BuildOptions{}, crex.Wrap(ErrInvalidPath, err)
	}

	return es.BuildOptions{

		// We handle logging ourselves.
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
			plugin,
		},
	}, nil
}
