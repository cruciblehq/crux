package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux container update' command.
type ContainerUpdateCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	ID      string `arg:"" required:"" help:"Container identifier."`
	Path    string `required:"" help:"Path to the new OCI image tarball."`
}

// Stops the container, re-imports the image, and restarts.
func (c *ContainerUpdateCmd) Run(ctx context.Context) error {
	slog.Info("updating container...", "id", c.ID)

	client := daemon.NewClient()
	if err := client.ContainerUpdate(ctx, &protocol.ContainerUpdateRequest{
		Ref:     c.Ref,
		Version: c.Version,
		ID:      c.ID,
		Path:    c.Path,
	}); err != nil {
		return err
	}

	slog.Info("container updated")
	return nil
}
