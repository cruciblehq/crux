package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/compute"
)

// Represents the 'crux local stop' command.
type LocalStopCmd struct{}

// Stops the local environment.
func (c *LocalStopCmd) Run(ctx context.Context) error {
	slog.Info("stopping local environment...")

	p, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return err
	}
	name := compute.DefaultLocalMachineID()

	if err := p.Stop(ctx, name); err != nil {
		return localError(err, "cannot stop local environment", "Run 'crux local start' to start the local environment first.")
	}

	slog.Info("local environment stopped")
	return nil
}
