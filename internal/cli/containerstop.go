package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Represents the 'crux container stop' command.
type ContainerStopCmd struct {
	Ref []string `arg:"" required:"" help:"Crucible resource reference (e.g., my-namespace/my-service 1.0.0)."`
	ID  string   `name:"id" optional:"" help:"Container identifier. Defaults to the resource name."`
}

// Stops a running container's task.
func (c *ContainerStopCmd) Run(ctx context.Context) error {
	ref, err := reference.Parse(strings.Join(c.Ref, " "), resource.TypeService, nil)
	if err != nil {
		return err
	}

	ctr := runtime.NewContainer(ref.Identifier.Registry(), c.ID)

	slog.Info("stopping container...", "id", c.ID)

	if err := ctr.Stop(ctx); err != nil {
		return err
	}

	slog.Info("container stopped")
	return nil
}
