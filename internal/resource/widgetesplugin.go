package resource

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	es "github.com/evanw/esbuild/pkg/api"
)

const (

	// Identifier registered with esbuild for the internal module resolution plugin.
	esPluginName = "crux-internal-es-plugin"

	// Directory where third-party widget dependencies are resolved from
	// during esbuild module resolution.
	esDependenciesDirName = "node_modules"
)

// Internal esbuild plugin that handles module resolution logic via [resolveModule].
//
// Most dependencies are resolved from the local 'node_modules' directory (or
// whatever is set by DependenciesDirName), except for modules used by the
// Crucible stack, like react, react-reconciler, @cruciblehq/ui, and
// @cruciblehq/ui-web. These are treated as externals to avoid bundling them,
// since they are expected to be provided by the Crucible runtime.
var esplugin = es.Plugin{
	Name: esPluginName,
	Setup: func(build es.PluginBuild) {
		type buildState struct {
			startTime time.Time
		}

		state := &buildState{}

		// Log build start time
		build.OnStart(func() (es.OnStartResult, error) {
			state.startTime = time.Now()

			slog.Info("build started")

			return es.OnStartResult{}, nil
		})

		// Log build end
		build.OnEnd(func(result *es.BuildResult) (es.OnEndResult, error) {
			duration := time.Since(state.startTime)

			slog.Info(fmt.Sprintf("build finished in %s with %d error(s) and %d warning(s)",
				duration, len(result.Errors), len(result.Warnings)))

			return es.OnEndResult{}, nil
		})

		// Resolve modules
		build.OnResolve(es.OnResolveOptions{Filter: ".*"}, func(args es.OnResolveArgs) (es.OnResolveResult, error) {

			slog.Debug(fmt.Sprintf("resolving module '%s' in '%s'",
				args.Path, args.ResolveDir))

			return resolveEsModule(*build.InitialOptions, args)
		})
	},
}

// Delegates the resolution logic based on the kind of import.
func resolveEsModule(options es.BuildOptions, args es.OnResolveArgs) (es.OnResolveResult, error) {

	// Entry point
	if args.Kind == es.ResolveEntryPoint {
		return resolveEsEntryPoint(options, args)
	}

	// Everything else (import)
	return resolveEsImport(options, args)
}

// Handles the resolution of the build's entry point.
func resolveEsEntryPoint(_ es.BuildOptions, args es.OnResolveArgs) (es.OnResolveResult, error) {
	return es.OnResolveResult{
		Path: args.Path,
	}, nil
}

// Handles the resolution of imports.
//
// It treats certain modules as externals and resolves other imports from the
// local dependencies directory. It also handles relative and non-relative
// imports appropriately. External modules are marked to avoid bundling, while
// other modules are resolved to their actual file paths. Those modules are
// react, react-reconciler, @cruciblehq/ui, and @cruciblehq/ui-web.
func resolveEsImport(options es.BuildOptions, args es.OnResolveArgs) (es.OnResolveResult, error) {

	// Check if the import matches any external
	for _, e := range options.External {
		if args.Path == e || strings.HasPrefix(args.Path, e+"/") {
			return es.OnResolveResult{
				Path:        args.Path,
				External:    true,
				SideEffects: es.SideEffectsFalse,
			}, nil
		}
	}

	// Resolve from dependencies directory
	var path string

	// Relative imports are resolved normally
	if strings.HasPrefix(args.Path, "./") || strings.HasPrefix(args.Path, "../") {
		path = filepath.Join(args.ResolveDir, args.Path)
	} else {

		// Non-relative imports are resolved from the dependencies directory
		// Get project root from AbsWorkingDir (set by esbuild)
		projectRoot := options.AbsWorkingDir
		if projectRoot == "" {
			projectRoot, _ = os.Getwd()
		}
		widgetNodeModules := filepath.Join(projectRoot, esDependenciesDirName)
		path = filepath.Join(widgetNodeModules, args.Path)
	}

	// Check if path exists; if not, try adding .js or index.js
	if info, err := os.Lstat(path); err != nil {
		jsPath := path + ".js"
		if _, err := os.Lstat(jsPath); err == nil {
			path = jsPath
		}
		// If still doesn't exist, esbuild will handle the error
	} else if info.IsDir() {
		path = filepath.Join(path, "index.js")
	}

	return es.OnResolveResult{
		Path:        path,
		External:    false,
		SideEffects: es.SideEffectsTrue,
	}, nil
}
