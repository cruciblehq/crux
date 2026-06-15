package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
)

// Represents the 'crux local restart' command.
type LocalRestartCmd struct{}

// Stops and restarts the local environment, preserving state.
func (c *LocalRestartCmd) Run(ctx context.Context) error {
	slog.Info("restarting local environment...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	if err := b.Stop(ctx, name); err != nil {
		return err
	}

	if err := b.Start(ctx, name); err != nil {
		return err
	}

	slog.Info("local environment restarted")
	return nil
}
