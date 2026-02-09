package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container status' command.
type ContainerStatusCmd struct {
	Ref []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
	ID  string   `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
}

// Shows the current state of a container.
func (c *ContainerStatusCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	ctr := runtime.NewContainer(ref.Identifier.Registry(), c.ID)

	status, err := ctr.Status(ctx)
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}
