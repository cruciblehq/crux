package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux container exec' command.
//
// The command to execute follows after '--'.
type ContainerExecCmd struct {
	ID      string   `arg:"" required:"" help:"Container identifier."`
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to execute."`
}

// Executes a command inside a container and prints its output.
//
// The process exit code is propagated from the executed command.
func (c *ContainerExecCmd) Run(ctx context.Context) error {
	client := daemon.NewClient()
	result, err := client.ContainerExec(ctx, &protocol.ContainerExecRequest{
		ID:      c.ID,
		Command: c.Command,
	})
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
