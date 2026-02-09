package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container destroy' command.
type ContainerDestroyCmd struct {
	Ref []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
	ID  string   `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
}

// Removes a container and its snapshot from the runtime.
func (c *ContainerDestroyCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	ctr := runtime.NewContainer(ref.Identifier.Registry(), c.ID)

	slog.Info("destroying container...", "id", c.ID)

	if err := ctr.Destroy(ctx); err != nil {
		return err
	}

	slog.Info("container destroyed")
	return nil
}
