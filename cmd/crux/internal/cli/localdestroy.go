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

// Represents the 'crux local destroy' command.
type LocalDestroyCmd struct{}

// Destroys the local environment and all its data.
func (c *LocalDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying local environment...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	if err := b.Deprovision(ctx, name); err != nil {
		const description = "cannot destroy local environment"
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

	slog.Info("local environment destroyed")
	return nil
}
