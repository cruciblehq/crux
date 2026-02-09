package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux vm stop' command.
type VmStopCmd struct{}

// Executes the VM stop command.
func (c *VmStopCmd) Run(ctx context.Context) error {
	slog.Info("stopping vm...")

	if err := runtime.Stop(); err != nil {
		return err
	}

	slog.Info("vm stopped")
	return nil
}
