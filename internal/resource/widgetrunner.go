package resource

import (
	"context"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
	es "github.com/evanw/esbuild/pkg/api"
)

// [Runner] for Crucible widgets.
//
// Widgets are client-side JavaScript bundles built with esbuild. Only Build,
// Pack, and Push are supported; lifecycle operations (Start, Stop, Destroy,
// Exec, Status) return [ErrUnsupported].
type WidgetRunner struct {
	registry         string
	defaultNamespace string
}

// Returns a [WidgetRunner] configured with the given registry and namespace
// fallbacks for push operations.
func NewWidgetRunner(registry, defaultNamespace string) *WidgetRunner {
	return &WidgetRunner{
		registry:         registry,
		defaultNamespace: defaultNamespace,
	}
}

// Builds a Crucible widget based on the provided manifest.
//
// It converts the manifest options into esbuild build options, invokes
// esbuild to perform the build, and processes the build result to log
// messages and handle errors.
func (wr *WidgetRunner) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {

	// Correct manifest type?
	widget, ok := m.Config.(*manifest.Widget)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	// Convert to esbuild options
	esOptions, err := esBuildOptionsFromManifest(widget, output)
	if err != nil {
		return nil, err
	}

	// esbuild doesn't support context cancellation, so this is the last chance
	// to abort the build.
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Build with esbuild
	result := es.Build(esOptions)

	// Process build result
	if err := processEsBuildResult(result); err != nil {
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

func (wr *WidgetRunner) Start(_ context.Context, _ manifest.Manifest, _ string) error {
	return ErrUnsupported
}

func (wr *WidgetRunner) Stop(_ context.Context, _ manifest.Manifest) error {
	return ErrUnsupported
}

func (wr *WidgetRunner) Restart(_ context.Context, _ manifest.Manifest, _ string) error {
	return ErrUnsupported
}

func (wr *WidgetRunner) Reset(_ context.Context, _ manifest.Manifest, _ string) error {
	return ErrUnsupported
}

func (wr *WidgetRunner) Destroy(_ context.Context, _ manifest.Manifest) error {
	return ErrUnsupported
}

func (wr *WidgetRunner) Exec(_ context.Context, _ manifest.Manifest, _ []string) (*ExecResult, error) {
	return nil, ErrUnsupported
}

func (wr *WidgetRunner) Status(_ context.Context, _ manifest.Manifest) (*StatusResult, error) {
	return nil, ErrUnsupported
}

// Packages the widget's build output into a distributable archive.
//
// The dist directory must contain index.js.
func (wr *WidgetRunner) Pack(ctx context.Context, m manifest.Manifest, manifestPath, dist, output string) (*PackResult, error) {
	return pack(ctx, m, manifestPath, dist, output)
}

// Uploads a widget package archive to the Hub registry.
//
// packagePath must point to an archive created by [WidgetRunner.Pack].
func (wr *WidgetRunner) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, m, packagePath, wr.registry, wr.defaultNamespace)
}
