package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux image import' command.
type ImageImportCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	Path    string `required:"" help:"Path to the OCI image tarball."`
}

// Imports an OCI image tarball into the container runtime's image store.
func (c *ImageImportCmd) Run(ctx context.Context) error {
	opts, err := reference.NewIdentifierOptions(internal.DefaultRegistryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}
	id, err := reference.ParseIdentifier(c.Ref, "service", opts)
	if err != nil {
		return err
	}

	client, err := runtime.NewContainerdClient(id.Hostname())
	if err != nil {
		return err
	}
	defer client.Close()

	img := runtime.NewImage(client, id, c.Version)

	slog.Info("importing image...", "image", img)

	if err := img.Import(ctx, c.Path); err != nil {
		return err
	}

	slog.Info("image imported")
	return nil
}
