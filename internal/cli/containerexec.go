package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container exec' command.
type ContainerExecCmd struct {
	Ref     []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
	ID      string   `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
	Command []string `arg:"" required:"" passthrough:"" help:"Command and arguments to execute."`
}

// Executes a command inside a container and prints its output.
//
// The process exit code is propagated from the executed command.
func (c *ContainerExecCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	ctr := runtime.NewContainer(ref.Identifier.Registry(), c.ID)

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
