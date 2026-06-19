package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/utils-go/crex"
)

// Returns a restart-specific structured error when this command has enough
// context to improve user guidance.
func localRestartError(err error) error {
	const description = "cannot restart local environment"
	switch {
	case errors.Is(err, local.ErrHostNotCreated):
		return crex.UserError(description, "the local environment has not been provisioned").
			Recovery("Run 'crux local start' to provision the local environment first.").
			Reclassify(err)
	case errors.Is(err, local.ErrHostNotRunning):
		return crex.UserError(description, "the local environment is not running").
			Recovery("Run 'crux local start' instead of 'crux local restart'.").
			Reclassify(err)
	case errors.Is(err, local.ErrHostAlreadyProvisioned):
		return crex.UserError(description, "the local environment is already provisioned").
			Recovery("Run 'crux local start' to use the existing environment, or 'crux local reset' to recreate it.").
			Reclassify(err)
	default:
		return err
	}
}

// Represents the 'crux local restart' command.
type LocalRestartCmd struct{}

// Stops and restarts the local environment, preserving state.
func (c *LocalRestartCmd) Run(ctx context.Context) error {
	slog.Info("restarting local environment...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	if err := b.Stop(ctx, name); err != nil {
		return localRestartError(err)
	}

	if err := b.Start(ctx, name); err != nil {
		return localRestartError(err)
	}

	slog.Info("local environment restarted")
	return nil
}
