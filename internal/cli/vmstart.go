package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/vm"
)

// Represents the 'crux vm start' command.
type VmStartCmd struct{}

// Executes the VM start command.
func (c *VmStartCmd) Run(ctx context.Context) error {
	m, err := vm.NewMachine()
	if err != nil {
		return err
	}

	slog.Info("starting vm...")

	if err := m.Start(); err != nil {
		return err
	}

	slog.Info("vm started")
	return nil
}
