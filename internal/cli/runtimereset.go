package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/runtime"
)

// Represents the 'crux runtime reset' command.
type RuntimeResetCmd struct{}

// Destroys and recreates the container runtime environment from scratch.
func (c *RuntimeResetCmd) Run(ctx context.Context) error {
	slog.Info("resetting runtime...")

	if err := runtime.Reset(); err != nil {
		return err
	}

	slog.Info("runtime reset complete")
	return nil
}
