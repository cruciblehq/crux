package cli

import (
	"context"
	"errors"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
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
		if errors.Is(err, manifest.ErrServiceNotFound) {
			return crex.UserError("service not found", c.ID).
				Recovery("Use 'crux local list' to see the registered services.").
				Cause(err).
				Err()
		}
		return err
	}
	return nil
}
