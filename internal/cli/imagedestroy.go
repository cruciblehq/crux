package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux image destroy' command.
type ImageDestroyCmd struct {
	Ref []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
}

// Removes an image and all its containers from the runtime.
func (c *ImageDestroyCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	id := ref.Identifier
	img := runtime.NewImage(&id, ref.Version().String())

	slog.Info("destroying image...", "image", img)

	if err := img.Destroy(ctx); err != nil {
		return err
	}

	slog.Info("image destroyed")
	return nil
}
