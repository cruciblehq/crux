package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/compute"
)

// Represents the 'crux local status' command.
type LocalStatusCmd struct{}

// Shows the current state of the local environment.
func (c *LocalStatusCmd) Run(ctx context.Context) error {
	p, err := compute.ProviderFor(compute.Local)
	if err != nil {
		return err
	}
	name := compute.DefaultLocalMachineID()

	state, err := p.Status(ctx, name)
	if err != nil {
		return localError(err, "cannot read local environment status", "Run 'crux local start' to start the local environment first.")
	}

	fmt.Println(state)
	return nil
}
