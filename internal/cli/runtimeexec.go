package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux runtime exec' command.
type RuntimeExecCmd struct {
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to run inside the runtime."`
}

// Executes a command inside the runtime and prints its output.
//
// The process exit code is propagated from the executed command.
func (c *RuntimeExecCmd) Run(ctx context.Context) error {
	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.InstanceName
	result, err := b.Exec(ctx, name, c.Command[0], c.Command[1:]...)
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
