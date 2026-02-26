package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
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
		if errors.Is(err, daemon.ErrConnectionRefused) {
			return crex.SystemError("daemon connection refused", err.Error()).
				Fallback("Wait a few seconds and try again. If the problem persists, try 'crux runtime restart' or 'crux runtime reset'.").
				Cause(err).
				Err()
		}
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
