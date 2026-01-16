package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/build"
	"github.com/cruciblehq/crux/pkg/watch"
)

// Represents the 'crux build' command.
type BuildCmd struct {
	Watch bool `short:"w" help:"Watch for changes and rebuild automatically."` // Whether to watch for file changes
}

// Executes the build command.
//
// Performs an initial build, and if the Watch flag is set, enters watch mode.
func (c *BuildCmd) Run(ctx context.Context) error {

	// Build first (don't wait for changes)
	if err := build.Build(ctx); err != nil {
		return err
	}

	slog.Info("build completed successfully")

	// Watch mode
	if c.Watch {
		slog.Info("watching for changes (CTRL+C to exit)...")
		return c.watchAndRebuild(ctx)
	}

	return nil
}

// Watches for file changes and triggers rebuilds.
//
// Sets up a recursive file watcher on the current directory, listening for any
// changes. When a file change is detected, it triggers a rebuild. The function
// continues to watch for changes until the provided context is canceled.
func (c *BuildCmd) watchAndRebuild(ctx context.Context) error {
	callback := func(we *watch.WatchEvent) error {

		// Check for cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		slog.Info("change detected, rebuilding...", "file", we.Path)

		// Rebuild
		if err := build.Build(ctx); err != nil {
			slog.Error(err.Error())
			return nil
		}

		slog.Info("rebuild completed successfully")
		return nil
	}

	// Watch
	if _, err := watch.WatchRecursive(".", callback); err != nil {
		return err
	}

	// Wait for cancellation
	<-ctx.Done()
	return ctx.Err()
}
