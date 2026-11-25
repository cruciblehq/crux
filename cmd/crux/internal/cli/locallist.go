package cli

import (
	"context"
	"fmt"
)

// Represents the 'crux local list' command.
type LocalListCmd struct{}

// Lists all services registered in the local blueprint.
//
// Prints each service ID followed by its path, or by its registry reference
// when it has no path.
func (c *LocalListCmd) Run(_ context.Context) error {
	bp, err := localBlueprint()
	if err != nil {
		return err
	}

	for _, s := range bp.Services {
		target := s.Ref
		if s.Path != "" {
			target = s.Path
		}
		fmt.Printf("%s %s\n", s.ID, target)
	}
	return nil
}
