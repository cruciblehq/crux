package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux runtime destroy' command.
type RuntimeDestroyCmd struct{}

// Destroys the cruxd runtime instance and all its data.
func (c *RuntimeDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying runtime...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.InstanceName

	if err := b.Deprovision(ctx, name); err != nil {
		return err
	}

	slog.Info("runtime destroyed")
	return nil
}
