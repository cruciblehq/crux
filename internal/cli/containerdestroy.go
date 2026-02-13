package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container destroy' command.
type ContainerDestroyCmd struct {
	Ref     string `arg:"" required:"" help:"Resource path (e.g., my-namespace/my-service)."`
	Version string `arg:"" required:"" help:"Resource version (e.g., 1.0.0)."`
	ID      string `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
}

// Removes a container and its snapshot from the runtime.
func (c *ContainerDestroyCmd) Run(ctx context.Context) error {
	opts, err := reference.NewIdentifierOptions(internal.DefaultRegistryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}
	id, err := reference.ParseIdentifier(c.Ref, resource.TypeService, opts)
	if err != nil {
		return err
	}

	client, err := runtime.NewContainerdClient(id.Hostname())
	if err != nil {
		return err
	}
	defer client.Close()

	ctr := runtime.NewContainer(client, id.Hostname(), c.ID)

	slog.Info("destroying container...", "id", c.ID)

	if err := ctr.Destroy(ctx); err != nil {
		return err
	}

	slog.Info("container destroyed")
	return nil
}
