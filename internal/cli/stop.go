package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux stop' command.
type StopCmd struct{}

// Stops the resource.
func (c *StopCmd) Run(ctx context.Context) error {

	slog.Info("stopping resource...")

	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	if err := r.Stop(ctx, *man); err != nil {
		return err
	}

	slog.Info("resource stopped")
	return nil
}
