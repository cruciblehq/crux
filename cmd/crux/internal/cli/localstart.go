package cli

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/utils-go/crex"
)

// Returns a start-specific structured error when this command has enough
// context to improve user guidance.
func localStartError(err error) error {
	const description = "cannot start local environment"
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
	registryURL, err := url.Parse(internal.DefaultRegistryURL)
	if err != nil {
		return crex.ProgrammingError("invalid default registry URL", "the compiled-in default registry URL could not be parsed").
			Cause(err).
			Err()
	}

	state, err := b.Compute().Status(ctx, name)
	if err != nil {
		return localStartError(err)
	}

	switch state {
	case compute.StateNotProvisioned:
		imagePath, err := local.EnsureMachineImage(ctx, registryURL)
		if err != nil {
			return err
		}
		f, err := os.Open(imagePath)
		if err != nil {
			return crex.SystemError("cannot access machine image", "the machine image file is missing or unreadable").
				Recovery("Run 'crux local reset' to download the default image again.").
				Cause(err).
				Err()
		}
		defer f.Close()
		imageID, err := b.Compute().UploadImage(ctx, f)
		if err != nil {
			return localStartError(err)
		}
		if _, err := b.Compute().Provision(ctx, imageID, localPlanOptions()); err != nil {
			return localStartError(err)
		}
	case compute.StateStopped:
		if err := b.Compute().Start(ctx, name); err != nil {
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
