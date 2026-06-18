package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/crux/crex"
)

// Returns a reset-specific structured error when this command has enough
// context to improve user guidance.
func localResetError(err error) error {
	const description = "cannot reset local environment"
	switch {
	case errors.Is(err, local.ErrHostNotCreated):
		return crex.UserError(description, "the local environment has not been provisioned").
			Recovery("Run 'crux local start' to provision the local environment first.").
			Reclassify(err)
	case errors.Is(err, local.ErrHostNotRunning):
		return crex.UserError(description, "the local environment is not running").
			Recovery("Run 'crux local start' to start the local environment first.").
			Reclassify(err)
	case errors.Is(err, local.ErrHostAlreadyProvisioned):
		return crex.UserError(description, "the local environment is already provisioned").
			Recovery("Run 'crux local start' to use the existing environment, or 'crux local reset' to recreate it.").
			Reclassify(err)
	default:
		return err
	}
}

// Represents the 'crux local reset' command.
type LocalResetCmd struct{}

// Deprovisions and re-provisions the local environment from scratch.
func (c *LocalResetCmd) Run(ctx context.Context) error {
	slog.Info("resetting local environment...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}

	name := internal.DefaultInstanceName

	if err := b.Deprovision(ctx, name); err != nil {
		return localResetError(err)
	}

	imagePath, err := local.EnsureMachineImage(ctx)
	if err != nil {
		return localResetError(err)
	}
	f, err := os.Open(imagePath)
	if err != nil {
		return crex.SystemError("cannot access machine image", "the machine image file is missing or unreadable").
			Recovery("Check your network connection and try again.").
			Cause(err).
			Err()
	}
	defer f.Close()
	imageID, err := b.Upload(ctx, f)
	if err != nil {
		return localResetError(err)
	}
	if err := b.Provision(ctx, imageID, name, localPlanOptions()); err != nil {
		return localResetError(err)
	}

	slog.Info("local environment reset complete")
	return nil
}
