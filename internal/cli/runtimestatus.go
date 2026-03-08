package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux runtime status' command.
type RuntimeStatusCmd struct{}

// Shows the current state of the cruxd runtime instance.
func (c *RuntimeStatusCmd) Run(ctx context.Context) error {
	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.InstanceName
	state, err := b.Status(ctx, name)
	if err != nil {
		return err
	}

	fmt.Println(state)
	return nil
}
