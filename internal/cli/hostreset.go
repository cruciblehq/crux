package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux host reset' command.
type HostResetCmd struct{}

// Deprovisions and re-provisions the cruxd host instance from scratch.
func (c *HostResetCmd) Run(ctx context.Context) error {
	slog.Info("resetting host...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}

	name := internal.DefaultInstanceName
	if err := b.Deprovision(ctx, name); err != nil {
		return err
	}

	source, err := resource.NewSource(internal.DefaultRegistryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}
	if err := b.Provision(ctx, name, source); err != nil {
		return err
	}

	slog.Info("host reset complete")
	return nil
}
