package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux host stop' command.
type HostStopCmd struct{}

// Stops the compute host.
func (c *HostStopCmd) Run(ctx context.Context) error {
	slog.Info("stopping host...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	if err := b.Stop(ctx, name); err != nil {
		return err
	}

	slog.Info("host stopped")
	return nil
}
