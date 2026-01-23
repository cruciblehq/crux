package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/pkg/push"
)

// Represents the 'crux push' command.
type PushCmd struct {
	Registry string `help:"Hub registry URL (default: http://hub.cruciblehq.xyz:8080)."`
	Resource string `arg:"" help:"Resource to push (namespace/name)."`
}

// Executes the push command.
func (c *PushCmd) Run(ctx context.Context) error {
	registry := c.Registry
	if registry == "" {
		registry = internal.DefaultRegistryURL
	}

	opts := push.PushOptions{
		Registry:     registry,
		Resource:     c.Resource,
		Manifestfile: internal.Manifestfile,
		Package:      internal.Package,
	}

	slog.Info("pushing package...", "resource", c.Resource, "registry", registry)

	if err := push.Push(ctx, opts); err != nil {
		return err
	}

	slog.Info("package pushed successfully", "resource", c.Resource)

	return nil
}
