package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
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
	name := internal.DefaultInstanceName

	if err := b.Deprovision(ctx, name); err != nil {
		return err
	}

	slog.Info("local environment destroyed")
	return nil
}
