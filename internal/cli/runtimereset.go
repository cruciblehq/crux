package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux runtime reset' command.
type RuntimeResetCmd struct{}

// Deprovisions and re-provisions the cruxd runtime instance from scratch.
func (c *RuntimeResetCmd) Run(ctx context.Context) error {
	slog.Info("resetting runtime...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.InstanceName

	if err := b.Deprovision(ctx, name); err != nil {
		return err
	}

	if err := b.Provision(ctx, &compute.Config{Name: name, Version: internal.CruxdVersion}); err != nil {
		return err
	}

	slog.Info("runtime reset complete")
	return nil
}
