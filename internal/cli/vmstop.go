package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/vm"
)

// Represents the 'crux vm stop' command.
type VmStopCmd struct{}

// Executes the VM stop command.
func (c *VmStopCmd) Run(ctx context.Context) error {
	m, err := vm.NewMachine()
	if err != nil {
		return err
	}

	slog.Info("stopping vm...")

	if err := m.Stop(); err != nil {
		return err
	}

	slog.Info("vm stopped")
	return nil
}
