package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crux/compute"
)

// Returns a user error for a failed 'reset' when the cause is a known
// local environment state.
func localResetError(err error) error {
	return localError(err, "cannot reset local environment", "Run 'crux local start' to start the local environment first.")
}

// Represents the 'crux local reset' command.
type LocalResetCmd struct{}

// Deprovisions the local machine and provisions it again from scratch.
func (c *LocalResetCmd) Run(ctx context.Context) error {
	slog.Info("resetting local environment...")

	p, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return err
	}

	name := compute.DefaultLocalMachineID()
	if err := p.Deprovision(ctx, name); err != nil && !errors.Is(err, compute.ErrMachineNotProvisioned) {
		return localResetError(err)
	}

	if err := provisionLocal(ctx, p, localResetError); err != nil {
		return err
	}

	slog.Info("local environment reset complete")
	return nil
}
