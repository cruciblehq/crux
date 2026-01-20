package cli

import (
	"context"

	"github.com/cruciblehq/crux/pkg/crex"
)

// Represents the 'crux scaffold' command.
type ScaffoldCmd struct {
	Reference string `arg:"" optional:"" help:"Reference in the template context."`
}

// Executes the scaffold command.
func (c *ScaffoldCmd) Run(ctx context.Context) error {
	return crex.ProgrammingError("cannot initialize new resource project", "scaffold command not implemented yet").
		Err()
}
