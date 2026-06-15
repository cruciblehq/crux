package cli

import (
	"context"

	"github.com/cruciblehq/crux/manifest"
)

// Represents the 'crux local remove' command.
type LocalRemoveCmd struct {
	ID string `arg:"" help:"ID of the service to remove."`
}

// Removes a service from the local blueprint.
func (c *LocalRemoveCmd) Run(ctx context.Context) error {
	if err := modifyLocalBlueprint(ctx, func(bp *manifest.Blueprint) error {
		return bp.RemoveService(c.ID)
	}); err != nil {
		return err
	}
	return nil
}
