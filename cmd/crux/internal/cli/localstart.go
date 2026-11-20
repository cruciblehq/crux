package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/utils-go/crex"
)

// Returns a user error for a failed 'start' when the cause is a known
// local environment state.
func localStartError(err error) error {
	return localError(err, "cannot start local environment", "Run 'crux local reset' to recreate the local environment.")
}

// Represents the 'crux local start' command.
type LocalStartCmd struct{}

// Provisions the local machine if needed and starts it.
func (c *LocalStartCmd) Run(ctx context.Context) error {
	slog.Info("starting local environment...")

	p, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return err
	}

	name := compute.DefaultLocalMachineID()
	state, err := p.Status(ctx, name)
	if err != nil {
		return localStartError(err)
	}

	switch state {
	case compute.StateNotProvisioned:
		if err := provisionLocal(ctx, p, localStartError); err != nil {
			return err
		}
	case compute.StateStopped:
		if err := p.Start(ctx, name); err != nil {
			return localStartError(err)
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
