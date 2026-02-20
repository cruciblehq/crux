package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux container destroy' command.
type ContainerDestroyCmd struct {
	ID string `arg:"" required:"" help:"Container identifier."`
}

// Removes a container and its snapshot from the runtime.
func (c *ContainerDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying container...", "id", c.ID)

	client := daemon.NewClient()
	if err := client.ContainerDestroy(ctx, &protocol.ContainerDestroyRequest{
		ID: c.ID,
	}); err != nil {
		return err
	}

	slog.Info("container destroyed")
	return nil
}
