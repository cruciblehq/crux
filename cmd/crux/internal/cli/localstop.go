package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
)

// Represents the 'crux local stop' command.
type LocalStopCmd struct{}

// Stops the local environment.
func (c *LocalStopCmd) Run(ctx context.Context) error {
	slog.Info("stopping local environment...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	if err := b.Stop(ctx, name); err != nil {
		return err
	}

	slog.Info("local environment stopped")
	return nil
}
