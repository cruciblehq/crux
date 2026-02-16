package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/build"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/kit/watch"
	"github.com/cruciblehq/crux/paths"
)

// Represents the 'crux build' command.
type BuildCmd struct {
	Watch    bool   `short:"w" help:"Watch for changes and rebuild automatically."`
	Registry string `help:"Hub registry URL for fetching runtimes (default: http://hub.cruciblehq.xyz:8080)."`
}

// Executes the build command.
//
// Performs an initial build, and if the Watch flag is set, enters watch mode.
func (c *BuildCmd) Run(ctx context.Context) error {

	slog.Info("building resource...", "watch", c.Watch)

	registry := c.Registry
	if registry == "" {
		registry = internal.DefaultRegistryURL
	}

	// Build first (don't wait for changes)
	result, err := build.Build(ctx, build.Options{
		Manifest:         paths.Manifest(RootCmd.Context),
		Output:           paths.BuildDir(RootCmd.Context),
		Registry:         registry,
		DefaultNamespace: internal.DefaultNamespace,
	})
	if err != nil {
		return err
	}

	slog.Info("build completed successfully", "output", result.Output)

	// Watch mode
	if c.Watch {
		slog.Info("watching for changes (CTRL+C to exit)...")
		return c.watchAndRebuild(ctx, registry)
	}

	return nil
}

// Watches for file changes and triggers rebuilds.
//
// Sets up a recursive file watcher on the current directory, listening for any
// changes. When a file change is detected, it triggers a rebuild. The function
// continues to watch for changes until the provided context is canceled.
func (c *BuildCmd) watchAndRebuild(ctx context.Context, registry string) error {
	callback := func(we *watch.WatchEvent) error {

		// Check for cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		slog.Info("change detected, rebuilding...", "file", we.Path)

		// Rebuild
		if _, err := build.Build(ctx, build.Options{
			Manifest:         paths.Manifest(RootCmd.Context),
			Output:           paths.BuildDir(RootCmd.Context),
			Registry:         registry,
			DefaultNamespace: internal.DefaultNamespace,
		}); err != nil {
			slog.Error(err.Error())
			return nil
		}

		slog.Info("rebuild completed successfully")

		return nil
	}

	// Watch
	if _, err := watch.WatchRecursive(RootCmd.Context, callback); err != nil {
		return err
	}

	// Wait for cancellation
	<-ctx.Done()
	return ctx.Err()
}


