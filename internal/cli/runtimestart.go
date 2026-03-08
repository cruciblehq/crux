package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux runtime start' command.
type RuntimeStartCmd struct{}

// Provisions and starts the cruxd runtime instance.
func (c *RuntimeStartCmd) Run(ctx context.Context) error {
	slog.Info("starting runtime...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.InstanceName

	state, err := b.Status(ctx, name)
	if err != nil {
		return err
	}

	switch state {
	case compute.StateNotProvisioned:
		if err := b.Provision(ctx, &compute.Config{Name: name, Version: internal.CruxdVersion}); err != nil {
			return err
		}
	case compute.StateStopped:
		if err := b.Start(ctx, name); err != nil {
			return err
		}
	case compute.StateRunning:
		slog.Info("runtime started")
		return nil
	default:
		return fmt.Errorf("cannot start: runtime is %s", state)
	}

	slog.Info("runtime started")
	return nil
}
