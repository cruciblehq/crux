package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/protocol"
)

// Represents the 'crux image import' command.
type ImageImportCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	Path    string `required:"" help:"Path to the OCI image tarball."`
}

// Imports an OCI image tarball into the container runtime's image store.
func (c *ImageImportCmd) Run(ctx context.Context) error {
	slog.Info("importing image...", "ref", c.Ref, "version", c.Version)

	client := daemon.NewClient()
	if err := client.ImageImport(ctx, &protocol.ImageImportRequest{
		Ref:     c.Ref,
		Version: c.Version,
		Path:    c.Path,
	}); err != nil {
		return err
	}

	slog.Info("image imported")
	return nil
}
