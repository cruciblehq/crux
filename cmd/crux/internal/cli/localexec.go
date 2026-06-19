package cli

import (
	"context"
	"errors"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/utils-go/crex"
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
		const description = "cannot run command in local environment"
		switch {
		case errors.Is(err, local.ErrHostNotCreated):
			return crex.UserError(description, "the local environment has not been provisioned").
				Recovery("Run 'crux local start' to provision the local environment first.").
				Reclassify(err)
		case errors.Is(err, local.ErrHostNotRunning):
			return crex.UserError(description, "the local environment is not running").
				Recovery("Run 'crux local start' to start the local environment first.").
				Reclassify(err)
		case errors.Is(err, local.ErrHostAlreadyProvisioned):
			return crex.UserError(description, "the local environment is already provisioned").
				Recovery("Run 'crux local start' to use the existing environment, or 'crux local reset' to recreate it.").
				Reclassify(err)
		default:
			return err
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}
