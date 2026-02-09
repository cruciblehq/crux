package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux vm start' command.
type VmStartCmd struct{}

// Executes the VM start command.
func (c *VmStartCmd) Run(ctx context.Context) error {
	slog.Info("starting vm...")

	if err := runtime.Start(); err != nil {
		return err
	}

	slog.Info("vm started")
	return nil
}
