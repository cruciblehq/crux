package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/build"
	"github.com/cruciblehq/crux/pkg/manifest"
	"github.com/cruciblehq/crux/pkg/watch"
)

// Represents the 'crux build' command
type BuildCmd struct {
	Watch bool `short:"w" help:"Watch for changes and rebuild automatically."`
}

// Executes the build command
func (c *BuildCmd) Run(ctx context.Context) error {

	// Load manifest options
	man, err := manifest.Read()
	if err != nil {
		return err
	}

	// Build first (don't wait for changes)
	if err := build.Build(ctx, *man); err != nil {
		return err
	}

	slog.Info("build completed successfully")

	// Watch mode
	if c.Watch {
		slog.Info("watching for changes...")
		return c.watchAndRebuild(ctx, *man)
	}

	return nil
}

func (c *BuildCmd) watchAndRebuild(ctx context.Context, man manifest.Manifest) error {
	callback := func(we *watch.WatchEvent) error {
		// Check for cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		slog.Info("change detected, rebuilding...", "file", we.Path)

		if err := build.Build(ctx, man); err != nil {
			slog.Error("rebuild failed", "error", err)
			// Continue watching despite errors
			return nil
		}

		slog.Info("rebuild completed successfully")
		return nil
	}

	// TODO: Watch only relevant directories based on resource type
	if _, err := watch.WatchRecursive(".", callback); err != nil {
		return err
	}

	// Wait for cancellation
	<-ctx.Done()
	return ctx.Err()
}
