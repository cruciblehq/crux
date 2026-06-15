package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/source"
)

// Represents the 'crux cache remove' command.
type CacheRemoveCmd struct {
	Type       string   `arg:"" help:"Resource type (widget, service)."`
	References []string `arg:"" help:"References to remove (e.g., namespace/resource 1.0.0)."`
}

// Executes the cache remove command.
func (c *CacheRemoveCmd) Run(_ context.Context) error {
	slog.Info("removing cache entries...")

	resType, err := manifest.ParseResourceType(c.Type)
	if err != nil {
		return crex.UserError("invalid resource type", c.Type).
			Fallback("Use a valid resource type such as 'widget' or 'service'.").
			Err()
	}

	src, err := source.NewSource(internal.DefaultRegistryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}

	localCache, err := cache.Open()
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
func removeReference(c *cache.Cache, source source.Source, resType manifest.ResourceType, refStr string) error {
	ref, err := source.Parse(string(resType), refStr)
	if err != nil {
		return crex.UserError("invalid reference", err.Error()).
			Fallback("Use the format 'namespace/resource version'.").
			Err()
	}

	if ref.IsVersionBased() {
		return removeVersion(c, ref)
	}

	return removeAllVersions(c, ref)
}

// Removes a specific version from the cache.
func removeVersion(c *cache.Cache, ref *reference.Reference) error {
	_, err := c.Get(ref.Namespace(), ref.Name(), ref.Version().String())
	if errors.Is(err, cache.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return c.Delete(ref.Namespace(), ref.Name(), ref.Version().String())
}

// Removes all versions of a resource from the cache.
func removeAllVersions(c *cache.Cache, ref *reference.Reference) error {
	entries, err := c.List()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Namespace == ref.Namespace() && entry.Resource == ref.Name() {
			if err := c.Delete(entry.Namespace, entry.Resource, entry.String); err != nil {
				return err
			}
		}
	}

	return nil
}
