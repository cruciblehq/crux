package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux vm destroy' command.
type VmDestroyCmd struct{}

// Executes the VM destroy command.
func (c *VmDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying vm...")

	if err := runtime.Destroy(); err != nil {
		return err
	}

	slog.Info("vm destroyed")
	return nil
}
