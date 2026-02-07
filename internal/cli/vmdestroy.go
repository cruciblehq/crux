package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/vm"
)

// Represents the 'crux vm destroy' command.
type VmDestroyCmd struct{}

// Executes the VM destroy command.
func (c *VmDestroyCmd) Run(ctx context.Context) error {
	m, err := vm.NewMachine()
	if err != nil {
		return err
	}

	slog.Info("destroying vm...")

	if err := m.Destroy(); err != nil {
		return err
	}

	slog.Info("vm destroyed")
	return nil
}
