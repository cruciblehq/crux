package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/push"
)

// Pushes a resource package to the Hub registry.
type PushCmd struct {
	Hub      string `help:"Hub registry URL." default:"http://hub.cruciblehq.xyz:8080"`
	Resource string `arg:"" help:"Resource to push (namespace/name)."`
}

// Executes the push command.
func (cmd *PushCmd) Run(ctx context.Context) error {
	opts := push.PushOptions{
		HubURL:   cmd.Hub,
		Resource: cmd.Resource,
	}

	if err := push.Push(ctx, opts); err != nil {
		return err
	}

	slog.Info("package pushed successfully", "resource", cmd.Resource)

	return nil
}
