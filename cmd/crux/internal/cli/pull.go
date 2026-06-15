package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/registry"
)

// Represents the 'crux pull' command.
type PullCmd struct {
	Registry  string   `help:"Hub registry URL (default: http://hub.cruciblehq.xyz:8080)."`
	Type      string   `arg:"" help:"Resource type (widget, service)."`
	Reference []string `arg:"" help:"Resource reference (e.g., crucible/login 1.0.0)."`
}

// Executes the pull command.
func (c *PullCmd) Run(ctx context.Context) error {
	registryURL := c.Registry
	if registryURL == "" {
		registryURL = internal.DefaultRegistryURL
	}

	resType, err := manifest.ParseResourceType(c.Type)
	if err != nil {
		return crex.UserError("invalid resource type", c.Type).
			Fallback("Use a valid resource type such as 'widget' or 'service'.").
			Err()
	}

	src, err := registry.NewSource(registryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}

	raw := strings.Join(c.Reference, " ")

	slog.Info("pulling resource...",
		"reference", raw,
		"registry", registryURL,
	)

	ref, err := src.Parse(string(resType), raw)
	if err != nil {
		return err
	}

	result, err := src.Pull(ctx, ref)
	if err != nil {
		return err
	}

	slog.Info("resource pulled",
		"namespace", result.Namespace,
		"resource", result.Resource,
		"version", result.Version,
		"digest", result.Digest,
		"size", result.Size,
	)

	return nil
}
