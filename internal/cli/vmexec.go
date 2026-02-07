package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cruciblehq/crux/vm"
)

// Represents the 'crux vm exec' command.
type VmExecCmd struct {
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to run inside the VM."`
}

// Executes a command inside the VM and prints its output.
func (c *VmExecCmd) Run(ctx context.Context) error {
	m, err := vm.NewMachine()
	if err != nil {
		return err
	}

	result, err := m.Exec(c.Command[0], c.Command[1:]...)
	if err != nil {
		return err
	}

	if result.Stdout != "" {
		fmt.Fprint(os.Stdout, result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}
