package cli

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
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
		registry = internal.RegistryURL
	}

	resType, err := manifest.ParseResourceType(c.Type)
	if err != nil {
		return crex.UserError("invalid resource type", c.Type).
			Fallback("Use a valid resource type such as 'widget' or 'service'.").
			Err()
	}

	raw := strings.Join(c.Reference, " ")

	opts, err := reference.NewOptions(registry, internal.DefaultNamespace)
	if err != nil {
		return err
	}

	ref, err := reference.Parse(raw, string(resType), opts)
	if err != nil {
		return crex.UserError("invalid reference", "could not parse the resource reference").
			Fallback("Use the format 'namespace/resource version'.").
			Cause(err).
			Err()
	}

	slog.Info("pulling resource...", "reference", raw, "registry", ref.Registry())

	result, err := resource.Pull(ctx, ref)
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
