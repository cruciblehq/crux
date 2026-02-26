package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/runtime"
)

// Represents the 'crux runtime restart' command.
type RuntimeRestartCmd struct{}

// Stops and restarts the container runtime environment, preserving disk state.
func (c *RuntimeRestartCmd) Run(ctx context.Context) error {
	slog.Info("restarting runtime...")

	if err := runtime.Restart(); err != nil {
		return err
	}

	slog.Info("runtime restarted")
	return nil
}
