package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/protocol"
)

const argSeparator = "--"

// Strips a leading "--" argument separator from a command slice.
//
// Kong's passthrough tag includes the "--" in the captured arguments. This
// function removes it so the actual command is passed cleanly.
func stripArgSeparator(args []string) []string {
	if len(args) > 0 && args[0] == argSeparator {
		return args[1:]
	}
	return args
}

// Represents the 'crux exec' command.
type ExecCmd struct {
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to execute."`
}

// Executes a command inside the running resource.
func (c *ExecCmd) Run(ctx context.Context) error {
	man, err := resource.ReadManifest(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	if man.Resource.Type != manifest.TypeService {
		return resource.ErrUnsupported
	}

	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	client, err := b.Client(ctx, internal.DefaultInstanceName)
	if err != nil {
		return err
	}

	cmd := stripArgSeparator(c.Command)

	result, err := client.ContainerExec(ctx, &protocol.ContainerExecRequest{
		ID:      man.Resource.Name,
		Command: cmd,
	})
	if err != nil {
		if errors.Is(err, compute.ErrConnectionRefused) {
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
