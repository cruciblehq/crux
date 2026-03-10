package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux host restart' command.
type HostRestartCmd struct{}

// Stops and restarts the cruxd host instance, preserving state.
func (c *HostRestartCmd) Run(ctx context.Context) error {
	slog.Info("restarting host...")

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

	slog.Info("host restarted")
	return nil
}
