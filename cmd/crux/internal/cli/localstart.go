package cli

import (
	"context"
	"log/slog"
	"net/url"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/utils-go/crex"
)

// Returns a user error for a failed 'start' when the cause is a known
// local environment state.
func localStartError(err error) error {
	return localError(err, "cannot start local environment", "Run 'crux local start' to start the local environment first.")
}

// Represents the 'crux local start' command.
type LocalStartCmd struct{}

// Provisions and starts the local environment.
func (c *LocalStartCmd) Run(ctx context.Context) error {
	slog.Info("starting local environment...")

	p, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return err
	}

	name := compute.DefaultLocalMachineID()
	registryURL, err := url.Parse(internal.DefaultRegistryURL)
	if err != nil {
		return crex.ProgrammingError("invalid default registry URL", "the compiled-in default registry URL could not be parsed").
			Cause(err).
			Err()
	}

	state, err := p.Status(ctx, name)
	if err != nil {
		return localStartError(err)
	}

	switch state {
	case compute.StateNotProvisioned:
		imagePath, err := compute.EnsureMachineImage(ctx, registryURL)
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
		imageID, err := p.UploadImage(ctx, f)
		if err != nil {
			return localStartError(err)
		}
		if _, err := p.Provision(ctx, imageID, compute.ProvisionOptions{}); err != nil {
			return localStartError(err)
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
