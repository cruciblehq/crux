package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute"
)

// Represents the 'crux host status' command.
type HostStatusCmd struct{}

// Shows the current state of the cruxd host instance.
func (c *HostStatusCmd) Run(ctx context.Context) error {
	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	state, err := b.Status(ctx, name)
	if err != nil {
		return err
	}

	fmt.Println(state)
	return nil
}
