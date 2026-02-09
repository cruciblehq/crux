package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container update' command.
type ContainerUpdateCmd struct {
	Ref  []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
	ID   string   `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
	Path string   `required:"" help:"Path to the new OCI image tarball."`
}

// Stops the container, re-imports the image, and restarts.
func (c *ContainerUpdateCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	id := ref.Identifier
	img := runtime.NewImage(&id, ref.Version().String())
	ctr := runtime.NewContainer(id.Registry(), c.ID)

	slog.Info("updating container...", "id", c.ID)

	if err := img.Update(ctx, ctr, c.Path); err != nil {
		return err
	}

	slog.Info("container updated")
	return nil
}
