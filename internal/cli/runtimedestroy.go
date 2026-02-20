package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/runtime"
)

// Represents the 'crux runtime destroy' command.
type RuntimeDestroyCmd struct{}

// Destroys the container runtime environment and all its data.
func (c *RuntimeDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying runtime...")

	if err := runtime.Destroy(); err != nil {
		return err
	}

	slog.Info("runtime destroyed")
	return nil
}
