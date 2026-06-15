package cli

import (
	"context"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
)

// Represents the 'crux local exec' command.
type LocalExecCmd struct {
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to run inside the local environment."`
}

// Executes a command inside the local environment and prints its output.
//
// The process exit code is propagated from the executed command.
func (c *LocalExecCmd) Run(ctx context.Context) error {
	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	cmd := stripArgSeparator(c.Command)

	exitCode, err := b.Exec(ctx, name, os.Stdout, os.Stderr, cmd[0], cmd[1:]...)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}
