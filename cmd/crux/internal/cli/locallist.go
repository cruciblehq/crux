package cli

import (
	"context"
	"fmt"
)

// Represents the 'crux local list' command.
type LocalListCmd struct{}

// Lists all services registered in the local blueprint.
func (c *LocalListCmd) Run(_ context.Context) error {
	bp, err := localBlueprint()
	if err != nil {
		return err
	}

	for _, s := range bp.Services {
		fmt.Printf("%s %s\n", s.ID, s.Ref)
	}
	return nil
}
