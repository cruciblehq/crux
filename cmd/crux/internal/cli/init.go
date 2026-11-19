package cli

import (
	"context"

	"github.com/cruciblehq/utils-go/crex"
)

// Represents the 'crux init' command.
type InitCmd struct {
}

// Scaffolds a new resource in the working directory.
//
// Not implemented yet. Always returns an error.
func (c *InitCmd) Run(ctx context.Context) error {
	return crex.UserError("cannot scaffold resource", "crux init is not available yet").
		Recovery("Write a crucible.yaml manifest by hand, then run 'crux build'.").
		Err()
}
