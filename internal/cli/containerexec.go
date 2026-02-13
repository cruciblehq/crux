package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container exec' command.
//
// The command to execute follows after '--'.
type ContainerExecCmd struct {
	Ref     string   `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string   `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	ID      string   `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to execute."`
}

// Executes a command inside a container and prints its output.
//
// The process exit code is propagated from the executed command.
func (c *ContainerExecCmd) Run(ctx context.Context) error {
	opts, err := reference.NewIdentifierOptions(internal.DefaultRegistryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}
	id, err := reference.ParseIdentifier(c.Ref, resource.TypeService, opts)
	if err != nil {
		return err
	}

	client, err := runtime.NewContainerdClient(id.Hostname())
	if err != nil {
		return err
	}
	defer client.Close()

	ctr := runtime.NewContainer(client, id.Hostname(), c.ID)

	result, err := ctr.Exec(ctx, c.Command[0], c.Command[1:]...)
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
