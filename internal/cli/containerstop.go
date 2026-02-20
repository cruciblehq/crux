package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux container stop' command.
type ContainerStopCmd struct {
	ID string `arg:"" required:"" help:"Container identifier."`
}

// Stops a running container's task.
func (c *ContainerStopCmd) Run(ctx context.Context) error {
	slog.Info("stopping container...", "id", c.ID)

	client := daemon.NewClient()
	if err := client.ContainerStop(ctx, &protocol.ContainerStopRequest{
		ID: c.ID,
	}); err != nil {
		return err
	}

	slog.Info("container stopped")
	return nil
}
