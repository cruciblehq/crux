package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux runtime start' command.
type RuntimeStartCmd struct{}

// Provisions and starts the container runtime environment.
func (c *RuntimeStartCmd) Run(ctx context.Context) error {
	slog.Info("starting runtime...")

	if err := runtime.Start(); err != nil {
		return err
	}

	slog.Info("runtime started")
	return nil
}
