package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/push"
)

// Represents the 'crux push' command.
type PushCmd struct {
	Registry string `help:"Hub registry URL (default: http://hub.cruciblehq.xyz:8080)."`
}

// Executes the push command.
func (c *PushCmd) Run(ctx context.Context) error {
	registry := c.Registry
	if registry == "" {
		registry = internal.DefaultRegistryURL
	}

	opts := push.PushOptions{
		Registry:         registry,
		Manifestfile:     paths.Manifest(RootCmd.Context),
		Package:          paths.Package(RootCmd.Context),
		DefaultNamespace: internal.DefaultNamespace,
	}

	slog.Info("pushing package...", "registry", registry)

	if err := push.Push(ctx, opts); err != nil {
		return err
	}

	slog.Info("package pushed successfully")

	return nil
}
