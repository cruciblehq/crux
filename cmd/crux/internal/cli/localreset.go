package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/compute/local"
)

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
		return err
	}

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

	slog.Info("local environment reset complete")
	return nil
}
