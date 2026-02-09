package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux image import' command.
type ImageImportCmd struct {
	Ref  []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
	Path string   `required:"" help:"Path to the OCI image tarball."`
}

// Imports an OCI image tarball into the container runtime's image store.
func (c *ImageImportCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	id := ref.Identifier
	img := runtime.NewImage(&id, ref.Version().String())

	slog.Info("importing image...", "image", img)

	if err := img.Import(ctx, c.Path); err != nil {
		return err
	}

	slog.Info("image imported")
	return nil
}
