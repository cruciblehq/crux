package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux destroy' command.
type DestroyCmd struct{}

// Destroys the resource.
func (c *DestroyCmd) Run(ctx context.Context) error {

	slog.Info("destroying resource...")

	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	if err := r.Destroy(ctx, *man); err != nil {
		return err
	}

	slog.Info("resource destroyed")
	return nil
}
