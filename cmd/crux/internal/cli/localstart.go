package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/crux/crex"
)

// Represents the 'crux local start' command.
type LocalStartCmd struct{}

// Provisions and starts the local environment.
func (c *LocalStartCmd) Run(ctx context.Context) error {
	slog.Info("starting local environment...")

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
		imagePath, err := local.EnsureMachineImage(ctx)
		if err != nil {
			return err
		}
		f, err := os.Open(imagePath)
		if err != nil {
			return err
		}
		defer f.Close()
		imageID, err := b.Upload(ctx, f)
		if err != nil {
			return err
		}
		if err := b.Provision(ctx, imageID, name, localPlanOptions()); err != nil {
			return err
		}
	case compute.StateStopped:
		if err := b.Start(ctx, name); err != nil {
			return err
		}
	case compute.StateRunning:
		slog.Info("local environment already running")
		return nil
	default:
		return crex.Newf(ErrUnexpectedState, "cannot start, local environment is %s", state)
	}

	slog.Info("local environment started")
	return nil
}
