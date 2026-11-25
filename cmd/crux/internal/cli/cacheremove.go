package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/spec/registry/cache"
	"github.com/cruciblehq/utils-go/crex"
)

// Represents the 'crux cache remove' command.
type CacheRemoveCmd struct {
	Type       string   `arg:"" help:"Resource type (machine, runtime, service, interface, template, widget, or blueprint)."`
	References []string `arg:"" help:"References to remove (e.g., namespace/resource 1.0.0)."`
}

// Executes the cache remove command.
func (c *CacheRemoveCmd) Run(_ context.Context) error {
	slog.Info("removing cache entries...")

	resType, err := manifest.ParseResourceType(c.Type)
	if err != nil {
		return crex.UserError("invalid resource type", c.Type).
			Recovery("Use a valid resource type such as 'widget' or 'service'.").
			Err()
	}

	src, err := registry.NewSource(internal.DefaultRegistryURL, internal.DefaultNamespace, dirs.Registry())
	if err != nil {
		return err
	}

	localCache, err := cache.Open(dirs.Registry())
	if err != nil {
		return err
	}
	defer localCache.Close()

	for _, refStr := range c.References {
		if err := removeReference(localCache, src, resType, refStr); err != nil {
			return err
		}
	}

	slog.Info("cache entries removed")

	return nil
}

// Removes cache entries matching a reference.
func removeReference(c *cache.Cache, src registry.Source, resType manifest.ResourceType, refStr string) error {
	ref, err := src.Parse(string(resType), refStr)
	if err != nil {
		return crex.UserError("invalid reference", err.Error()).
			Recovery("Use the format 'namespace/resource version'.").
			Err()
	}

	return removeMatchingVersions(c, ref)
}

// Removes the cached versions of a resource that the reference's version
// constraint matches.
func removeMatchingVersions(c *cache.Cache, ref *reference.Reference) error {
	entries, err := c.List()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Namespace != ref.Namespace() || entry.Resource != ref.Name() {
			continue
		}
		matches, err := ref.Version().Matches(entry.String)
		if err != nil {
			return err
		}
		if !matches {
			continue
		}
		if err := c.Delete(entry.Namespace, entry.Resource, entry.String); err != nil {
			return err
		}
	}

	return nil
}
