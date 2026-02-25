package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
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

	slog.Info("pushing package...", "registry", registry)

	man, r, err := resource.Resolve(paths.Manifest(RootCmd.Context), resource.Options{
		DefaultRegistry:  registry,
		DefaultNamespace: internal.DefaultNamespace,
	})
	if err != nil {
		return err
	}

	if err := r.Push(ctx, *man, paths.Package(RootCmd.Context)); err != nil {
		return err
	}

	slog.Info("package pushed successfully")

	return nil
}
