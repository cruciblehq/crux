package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/runtime"
)

// Represents the 'crux runtime stop' command.
type RuntimeStopCmd struct{}

// Stops the container runtime environment.
func (c *RuntimeStopCmd) Run(ctx context.Context) error {
	slog.Info("stopping runtime...")

	if err := runtime.Stop(); err != nil {
		return err
	}

	slog.Info("runtime stopped")
	return nil
}
