package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Represents the 'crux status' command.
type StatusCmd struct{}

// Shows the current state of the resource.
func (c *StatusCmd) Run(ctx context.Context) error {
	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context))
	if err != nil {
		return err
	}

	result, err := r.Status(ctx, *man)
	if err != nil {
		return err
	}

	fmt.Println(result.Status)
	return nil
}
