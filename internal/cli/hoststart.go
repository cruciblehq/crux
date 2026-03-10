package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux host start' command.
type HostStartCmd struct{}

// Provisions and starts the cruxd host instance.
func (c *HostStartCmd) Run(ctx context.Context) error {
	slog.Info("starting host...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	state, err := b.Status(ctx, name)
	if err != nil {
		return err
	}

	switch state {
	case compute.StateNotProvisioned:
		source, err := resource.NewSource(internal.DefaultRegistryURL, internal.DefaultNamespace)
		if err != nil {
			return err
		}
		if err := b.Provision(ctx, name, source); err != nil {
			return err
		}
	case compute.StateStopped:
		if err := b.Start(ctx, name); err != nil {
			return err
		}
	case compute.StateRunning:
		slog.Info("host already running")
		return nil
	default:
		return crex.Wrapf(ErrUnexpectedState, "cannot start: host is %s", state)
	}

	slog.Info("host started")
	return nil
}
