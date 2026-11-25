package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
	"github.com/cruciblehq/utils-go/watch"
)

// Represents the 'crux build' command.
type BuildCmd struct {
	Watch       bool   `short:"w" help:"Watch for changes and rebuild automatically."`
	Registry    string `help:"Registry to fetch resources from (default: the one set with 'crux registry use')."`
	Environment string `short:"e" help:"Blueprint environment to build against." default:""`
}

// Performs an initial build, and if the Watch flag is set, enters watch mode.
func (c *BuildCmd) Run(ctx context.Context) error {

	slog.Info("building resource...", "watch", c.Watch)

	src, err := openRegistrySource(c.Registry)
	if err != nil {
		return err
	}

	// Build first (don't wait for changes)
	result, err := c.build(ctx, src)
	if err != nil {
		return err
	}

	slog.Info("build completed successfully", "output", result.Output)

	// Watch mode
	if c.Watch {
		slog.Info("watching for changes (CTRL+C to exit)...")
		return c.watchAndRebuild(ctx, src)
	}

	return nil
}

// Watches for file changes and triggers rebuilds.
//
// Sets up a recursive file watcher on the current directory, listening for any
// changes. When a file change is detected, it triggers a rebuild. The function
// continues to watch for changes until the provided context is canceled.
func (c *BuildCmd) watchAndRebuild(ctx context.Context, src registry.Source) error {
	callback := func(we *watch.Event) error {

		// Check for cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		slog.Info("change detected, rebuilding...", "file", we.Path)

		if _, err := c.build(ctx, src); err != nil {
			crex.LogError(slog.Default(), err)
			return nil
		}

		slog.Info("rebuild completed successfully")

		return nil
	}

	if _, err := watch.WatchRecursive(RootCmd.Context, callback); err != nil {
		return crex.SystemError("cannot watch for changes", "the file watcher could not be started").
			Recovery("Try running the command again without watch mode.").
			Cause(err).
			Err()
	}

	// Wait for cancellation
	<-ctx.Done()
	return ctx.Err()
}

// Reads the manifest, creates the output directory, and builds the resource.
func (c *BuildCmd) build(ctx context.Context, src registry.Source) (*resource.BuildResult, error) {
	man, err := readProjectManifest()
	if err != nil {
		return nil, err
	}

	output := resource.BuildDir(RootCmd.Context)
	if err := os.MkdirAll(output, file.DefaultDirMode); err != nil {
		return nil, crex.SystemError("cannot prepare build output", "failed to create the build output directory").
			Recoveryf("Make sure you have write access to %s, then try again.", output).
			Cause(crex.Wrap(resource.ErrFileSystemOperation, err)).
			Err()
	}

	result, err := resource.Build(ctx, *man, src, RootCmd.Context, c.Environment, output)
	if err != nil {
		switch {
		case errors.Is(err, resource.ErrUnsupportedType):
			return nil, crex.UserError("cannot build resource", "the resource type is not supported by the build command").
				Recovery("Use a buildable type.").
				Reclassify(err)
		default:
			return nil, err
		}
	}

	return result, nil
}

// Reads the manifest of the project in the directory set by -C.
//
// Returns a user error when the manifest is missing or cannot be parsed.
func readProjectManifest() (*manifest.Manifest, error) {
	man, err := resource.ReadManifest(RootCmd.Context)
	if errors.Is(err, os.ErrNotExist) {
		return nil, crex.UserError("no manifest found", "no crucible.yaml was found in the current directory").
			Recovery("Run this command from a directory containing a crucible.yaml.").
			Cause(err).
			Err()
	}
	if err != nil {
		return nil, crex.UserError("invalid manifest", "the crucible.yaml could not be read or parsed").
			Recovery("Check that crucible.yaml is present and valid.").
			Cause(err).
			Err()
	}
	return man, nil
}
