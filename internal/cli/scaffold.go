package cli

import "github.com/cruciblehq/crux/pkg/crex"

// Scaffolds a new Crucible resource.
type ScaffoldCmd struct {
	Reference string `arg:"" optional:"" help:"Reference in the template context."`
}

func (c *ScaffoldCmd) Run() error {
	return crex.ProgrammingError("cannot initialize new resource project", "scaffold command not implemented yet").
		Err()
}
