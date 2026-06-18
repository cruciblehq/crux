package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/crux/crex"
)

// Represents the 'crux local status' command.
type LocalStatusCmd struct{}

// Shows the current state of the local environment.
func (c *LocalStatusCmd) Run(ctx context.Context) error {
	b, err := compute.BackendFor(compute.Local)
	if err != nil {
		return err
	}
	name := internal.DefaultInstanceName

	state, err := b.Status(ctx, name)
	if err != nil {
		const description = "cannot read local environment status"
		switch {
		case errors.Is(err, local.ErrHostNotCreated):
			return crex.UserError(description, "the local environment has not been provisioned").
				Recovery("Run 'crux local start' to provision the local environment first.").
				Reclassify(err)
		case errors.Is(err, local.ErrHostNotRunning):
			return crex.UserError(description, "the local environment is not running").
				Recovery("Run 'crux local start' to start the local environment first.").
				Reclassify(err)
		case errors.Is(err, local.ErrHostAlreadyProvisioned):
			return crex.UserError(description, "the local environment is already provisioned").
				Recovery("Run 'crux local start' to use the existing environment, or 'crux local reset' to recreate it.").
				Reclassify(err)
		default:
			return err
		}
	}

	fmt.Println(state)
	return nil
}
