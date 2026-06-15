package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/registry"
	"github.com/cruciblehq/crux/resource"
)

// Represents the 'crux push' command.
type PushCmd struct {
	Registry string `help:"Hub registry URL (default: http://hub.cruciblehq.xyz:8080)."`
}

// Executes the push command.
func (c *PushCmd) Run(ctx context.Context) error {
	registryURL := c.Registry
	if registryURL == "" {
		registryURL = internal.DefaultRegistryURL
	}

	slog.Info("pushing package...", "registry", registryURL)

	src, err := registry.NewSource(registryURL, internal.DefaultNamespace)
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
