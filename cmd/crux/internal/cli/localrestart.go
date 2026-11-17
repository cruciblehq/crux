package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
)

// Returns a user error for a failed 'restart' when the cause is a known
// local environment state.
func localRestartError(err error) error {
	return localError(err, "cannot restart local environment", "Run 'crux local start' instead of 'crux local restart'.")
}

// Represents the 'crux local restart' command.
type LocalRestartCmd struct{}

// Stops and restarts the local environment, preserving state.
func (c *LocalRestartCmd) Run(ctx context.Context) error {
	slog.Info("restarting local environment...")

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	if err := b.Compute().Stop(ctx, name); err != nil {
		return localRestartError(err)
	}

	if err := b.Compute().Start(ctx, name); err != nil {
		return localRestartError(err)
	}

	slog.Info("local environment restarted")
	return nil
}
