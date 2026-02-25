package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux exec' command.
type ExecCmd struct {
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to execute."`
}

// Executes a command inside the running resource.
func (c *ExecCmd) Run(ctx context.Context) error {
	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	result, err := r.Exec(ctx, *man, c.Command)
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
