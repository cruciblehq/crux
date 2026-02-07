package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/vm"
)

// Represents the 'crux vm status' command.
type VmStatusCmd struct{}

// Executes the VM status command.
func (c *VmStatusCmd) Run(ctx context.Context) error {
	m, err := vm.NewMachine()
	if err != nil {
		return err
	}

	status, err := m.Status()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}
