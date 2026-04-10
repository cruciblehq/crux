package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/crux/internal/runtime"
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
	rt, err := b.Runtime(ctx, internal.DefaultInstanceName)
	if err != nil {
		return err
	}
	defer rt.Close()

	name := runtime.ContainerID(man.Resource.Name)
	cmd := stripArgSeparator(c.Command)

	result, err := rt.Container(name).ExecArgs(ctx, cmd)
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
