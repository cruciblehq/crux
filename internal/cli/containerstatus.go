package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux container status' command.
type ContainerStatusCmd struct {
	ID string `arg:"" required:"" help:"Container identifier."`
}

// Shows the current state of a container.
func (c *ContainerStatusCmd) Run(ctx context.Context) error {
	client := daemon.NewClient()
	result, err := client.ContainerStatus(ctx, &protocol.ContainerStatusRequest{
		ID: c.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println(result.Status)
	return nil
}
