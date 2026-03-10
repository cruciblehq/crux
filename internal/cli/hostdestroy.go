package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux host destroy' command.
type HostDestroyCmd struct{}

// Destroys the cruxd host instance and all its data.
func (c *HostDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying host...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	if err := b.Deprovision(ctx, name); err != nil {
		return err
	}

	slog.Info("host destroyed")
	return nil
}
