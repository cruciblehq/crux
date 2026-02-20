package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux image start' command.
type ImageStartCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	ID      string `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
}

// Starts a new container from the image.
func (c *ImageStartCmd) Run(ctx context.Context) error {
	slog.Info("starting container...", "ref", c.Ref, "version", c.Version, "id", c.ID)

	client := daemon.NewClient()
	if err := client.ImageStart(ctx, &protocol.ImageStartRequest{
		Ref:     c.Ref,
		Version: c.Version,
		ID:      c.ID,
	}); err != nil {
		return err
	}

	slog.Info("container started")
	return nil
}
