package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux runtime stop' command.
type RuntimeStopCmd struct{}

// Stops the cruxd runtime instance.
func (c *RuntimeStopCmd) Run(ctx context.Context) error {
	slog.Info("stopping runtime...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.InstanceName

	if err := b.Stop(ctx, name); err != nil {
		return err
	}

	slog.Info("runtime stopped")
	return nil
}
