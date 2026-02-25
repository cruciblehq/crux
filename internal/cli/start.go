package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux start' command.
type StartCmd struct{}

// Starts the resource.
func (c *StartCmd) Run(ctx context.Context) error {

	slog.Info("starting resource...")

	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	if err := r.Start(ctx, *man, paths.ImageTar(RootCmd.Context)); err != nil {
		return err
	}

	slog.Info("resource started")
	return nil
}
