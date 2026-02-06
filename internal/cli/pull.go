package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/pull"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/resource"
)

// Represents the 'crux pull' command.
type PullCmd struct {
	Registry  string   `help:"Hub registry URL (default: http://hub.cruciblehq.xyz:8080)."`
	Type      string   `arg:"" help:"Resource type (widget, service)."`
	Reference []string `arg:"" help:"Resource reference (e.g., crucible/login 1.0.0)."`
}

// Executes the pull command.
func (c *PullCmd) Run(ctx context.Context) error {
	registry := c.Registry
	if registry == "" {
		registry = internal.DefaultRegistryURL
	}

	resType, err := resource.ParseType(c.Type)
	if err != nil {
		return crex.UserError("invalid resource type", c.Type).
			Fallback("Use a valid resource type such as 'widget' or 'service'.").
			Err()
	}

	reference := strings.Join(c.Reference, " ")

	opts := pull.Options{
		Registry:  registry,
		Reference: reference,
		Type:      resType,
	}

	slog.Info("pulling resource...", "reference", reference, "registry", registry)

	result, err := pull.Pull(ctx, opts)
	if err != nil {
		return err
	}

	if result.Cached {
		slog.Info("resource already cached",
			"namespace", result.Namespace,
			"resource", result.Resource,
			"version", result.Version,
			"digest", result.Digest,
		)
	} else {
		slog.Info("resource pulled successfully",
			"namespace", result.Namespace,
			"resource", result.Resource,
			"version", result.Version,
			"digest", result.Digest,
			"size", result.Size,
		)
	}

	return nil
}
