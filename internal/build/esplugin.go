package build

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

	// Name of the internal esbuild plugin
	EsPluginName = "crux-internal-es-plugin"

	// Name of the dependencies directory
	DependenciesDirName = "node_modules"
)

// [esplugin] is an internal esbuild plugin that handles module resolution
// logic via [resolveModule]. Most dependencies are resolved from the local
// 'node_modules' directory (or whatever is set by DependenciesDirName), except
// for modules used by the Crucible stack, like react, react-reconciler,
// @cruciblehq/ui, and @cruciblehq/ui-web. These are treated as externals to
// avoid bundling them, since they are expected to be provided by the Crucible
// runtime environment.
var esplugin = es.Plugin{
	Name: EsPluginName,
	Setup: func(build es.PluginBuild) {

		var timer time.Time

		// Log build start time
		build.OnStart(func() (es.OnStartResult, error) {

			timer = time.Now()

			slog.Info("build started")

			return es.OnStartResult{}, nil
		})

		// Log build end
		build.OnEnd(func(result *es.BuildResult) (es.OnEndResult, error) {

			duration := time.Since(timer)

			slog.Info(fmt.Sprintf("build finished in %s with %d error(s) and %d warning(s)",
				duration, len(result.Errors), len(result.Warnings)))

			return es.OnEndResult{}, nil
		})

		// Resolve modules
		build.OnResolve(es.OnResolveOptions{Filter: ".*"}, func(args es.OnResolveArgs) (es.OnResolveResult, error) {

			slog.Debug(fmt.Sprintf("resolving module '%s' in '%s'",
				args.Path, args.ResolveDir))

			return resolveModule(args)
		})
	},
}

// Delegates the resolution logic based on the kind of import.
func resolveModule(args es.OnResolveArgs) (es.OnResolveResult, error) {

	// Entry point
	if args.Kind == es.ResolveEntryPoint {
		return resolveEntryPoint(args)
	}

	// Everything else (import)
	return resolveImport(args)
}

// Handles the resolution of the build's entry point.
func resolveEntryPoint(_ es.OnResolveArgs) (es.OnResolveResult, error) {
	return es.OnResolveResult{}, nil
}

// Handles the resolution of imports.
//
// It treats certain modules as externals and resolves other imports from the
// local dependencies directory. It also handles relative and non-relative
// imports appropriately. External modules are marked to avoid bundling, while
// other modules are resolved to their actual file paths. Those modules are
// react, react-reconciler, @cruciblehq/ui, and @cruciblehq/ui-web.
func resolveImport(args es.OnResolveArgs) (es.OnResolveResult, error) {

	// Skip externals
	externals := []string{
		"react",
		"react-reconciler",
		"@cruciblehq/ui",
		"@cruciblehq/ui-web",
	}

	// Check if the import matches any external
	for _, e := range externals {
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
		wd, err := os.Getwd()
		if err != nil {
			return es.OnResolveResult{}, err
		}
		widgetNodeModules := filepath.Join(wd, DependenciesDirName)
		path = filepath.Join(widgetNodeModules, args.Path)
	}

	// Check if path exists; if not, try adding .js or index.js
	if info, err := os.Lstat(path); err != nil {
		path = path + ".js"
	} else if info.IsDir() {
		path = filepath.Join(path, "index.js")
	}

	return es.OnResolveResult{
		Path:        path,
		External:    false,
		SideEffects: es.SideEffectsTrue,
	}, nil
}
