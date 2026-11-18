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

// Returns a user error for a failed 'reset' when the cause is a known
// local environment state.
func localResetError(err error) error {
	return localError(err, "cannot reset local environment", "Run 'crux local start' to start the local environment first.")
}

// Represents the 'crux local reset' command.
type LocalResetCmd struct{}

// Deprovisions and re-provisions the local environment from scratch.
func (c *LocalResetCmd) Run(ctx context.Context) error {
	slog.Info("resetting local environment...")

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

	if err := p.Deprovision(ctx, name); err != nil {
		return localResetError(err)
	}

	imagePath, err := compute.EnsureMachineImage(ctx, registryURL)
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
	imageID, err := p.UploadImage(ctx, f)
	if err != nil {
		return localResetError(err)
	}
	if _, err := p.Provision(ctx, imageID, compute.ProvisionOptions{}); err != nil {
		return localResetError(err)
	}

	slog.Info("local environment reset complete")
	return nil
}
