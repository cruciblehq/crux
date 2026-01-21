package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/push"
)

// Represents the 'crux push' command.
type PushCmd struct {
	Hub      string `help:"Hub registry URL." default:"http://hub.cruciblehq.xyz:8080"`
	Resource string `arg:"" help:"Resource to push (namespace/name)."`
}

// Executes the push command.
func (c *PushCmd) Run(ctx context.Context) error {
	opts := push.PushOptions{
		HubURL:       c.Hub,
		Resource:     c.Resource,
		Manifestfile: Manifestfile,
		Package:      Package,
	}

	slog.Info("pushing package...", "resource", c.Resource, "hub", c.Hub)

	if err := push.Push(ctx, opts); err != nil {
		return err
	}

	slog.Info("package pushed successfully", "resource", c.Resource)

	return nil
}
