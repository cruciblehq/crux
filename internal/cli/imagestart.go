package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux image start' command.
type ImageStartCmd struct {
	Ref []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
	ID  string   `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
}

// Starts a new container from the image.
func (c *ImageStartCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	id := ref.Identifier
	img := runtime.NewImage(&id, ref.Version().String())

	slog.Info("starting container...", "image", img, "id", c.ID)

	ctr, err := img.Start(ctx, c.ID)
	if err != nil {
		return err
	}

	_ = ctr
	slog.Info("container started")
	return nil
}
