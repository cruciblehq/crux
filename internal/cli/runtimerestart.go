package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux runtime restart' command.
type RuntimeRestartCmd struct{}

// Stops and restarts the cruxd runtime instance, preserving state.
func (c *RuntimeRestartCmd) Run(ctx context.Context) error {
	slog.Info("restarting runtime...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.InstanceName

	if err := b.Stop(ctx, name); err != nil {
		return err
	}

	if err := b.Start(ctx, name); err != nil {
		return err
	}

	slog.Info("runtime restarted")
	return nil
}
