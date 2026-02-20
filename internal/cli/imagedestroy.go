package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux image destroy' command.
type ImageDestroyCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
}

// Removes an image and all its containers from the runtime.
func (c *ImageDestroyCmd) Run(ctx context.Context) error {
	slog.Info("destroying image...", "ref", c.Ref, "version", c.Version)

	client := daemon.NewClient()
	if err := client.ImageDestroy(ctx, &protocol.ImageDestroyRequest{
		Ref:     c.Ref,
		Version: c.Version,
	}); err != nil {
		return err
	}

	slog.Info("image destroyed")
	return nil
}
