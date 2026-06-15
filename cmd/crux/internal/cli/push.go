package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/source"
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

	src, err := source.NewSource(registry, internal.DefaultNamespace)
	if err != nil {
		return err
	}

	man, err := manifest.ReadAt(RootCmd.Context)
	if err != nil {
		return err
	}

	if err := resource.Push(ctx, src, *man, files.Package(RootCmd.Context)); err != nil {
		return err
	}

	slog.Info("package pushed successfully")

	return nil
}
