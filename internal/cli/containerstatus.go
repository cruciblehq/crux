package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container status' command.
type ContainerStatusCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	ID      string `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
}

// Shows the current state of a container.
func (c *ContainerStatusCmd) Run(ctx context.Context) error {
	opts, err := reference.NewIdentifierOptions(internal.DefaultRegistryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}
	id, err := reference.ParseIdentifier(c.Ref, "service", opts)
	if err != nil {
		return err
	}

	client, err := runtime.NewContainerdClient(id.Hostname())
	if err != nil {
		return err
	}
	defer client.Close()

	ctr := runtime.NewContainer(client, id.Hostname(), c.ID)

	status, err := ctr.Status(ctx)
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}
