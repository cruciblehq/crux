package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/compute"
)

// Represents the 'crux local destroy' command.
type LocalDestroyCmd struct{}

// Destroys the local environment and all its data.
func (c *LocalDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying local environment...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := compute.DefaultLocalInstanceID()

	if err := b.Compute().Deprovision(ctx, name); err != nil {
		return localError(err, "cannot destroy local environment", "Run 'crux local start' to start the local environment first.")
	}

	slog.Info("local environment destroyed")
	return nil
}
