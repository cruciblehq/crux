package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container update' command.
type ContainerUpdateCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	ID      string `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
	Path    string `required:"" help:"Path to the new OCI image tarball."`
}

// Stops the container, re-imports the image, and restarts.
func (c *ContainerUpdateCmd) Run(ctx context.Context) error {
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
	ctr := runtime.NewContainer(client, id.Hostname(), c.ID)

	slog.Info("updating container...", "id", c.ID)

	if err := img.Update(ctx, ctr, c.Path); err != nil {
		return err
	}

	slog.Info("container updated")
	return nil
}
