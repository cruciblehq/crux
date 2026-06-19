package cli

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/crex"
)

// Represents the 'crux pull' command.
type PullCmd struct {
	Registry  string   `help:"Hub registry URL (default: http://hub.cruciblehq.xyz:8080)."`
	Type      string   `arg:"" help:"Resource type (e.g., widget, service)."`
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
			Recovery("Use a valid resource type such as 'widget' or 'service'.").
			Err()
	}

	src, err := hub.NewSource(registryURL, internal.DefaultNamespace)
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
		return crex.UserError("invalid resource reference", raw).
			Recovery("Use a valid reference.").
			Cause(err).
			Err()
	}

	result, err := src.Pull(ctx, ref)
	if err != nil {
		const description = "cannot pull resource"
		switch {
		case errors.Is(err, registry.ErrNoVersions), errors.Is(err, registry.ErrNoMatchingVersion), errors.Is(err, registry.ErrTypeMismatch):
			return crex.UserError("resource not found", raw).
				Recovery("Check the resource reference and version, then try again.").
				Reclassify(err)
		case errors.Is(err, registry.ErrNoArchive):
			return crex.UserError(description, "the requested resource version has no uploaded archive").
				Recovery("Choose a different version, or republish the resource with an archive.").
				Reclassify(err)
		case errors.Is(err, registry.ErrResolveVersion):
			return crex.SystemError(description, "the registry could not resolve the requested resource version").
				Recovery("Check your network connection and try again.").
				Reclassify(err)
		case errors.Is(err, registry.ErrDownload):
			return crex.SystemError(description, "the resource archive could not be downloaded from the registry").
				Recovery("Check your network connection and try again.").
				Reclassify(err)
		case errors.Is(err, registry.ErrCacheOperation):
			return crex.SystemError(description, "the local cache could not store or extract the resource archive").
				Recovery("Run 'crux cache clear' and try again.").
				Reclassify(err)
		default:
			return err
		}
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
