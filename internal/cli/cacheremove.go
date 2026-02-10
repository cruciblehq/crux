package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
)

// Represents the 'crux cache remove' command.
type CacheRemoveCmd struct {
	References []string `arg:"" help:"References to remove (e.g., namespace/resource 1.0.0)."`
}

// Executes the cache remove command.
func (c *CacheRemoveCmd) Run(ctx context.Context) error {
	slog.Info("removing cache entries...")

	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return err
	}
	defer localCache.Close()

	for _, refStr := range c.References {
		if err := removeReference(ctx, localCache, refStr); err != nil {
			return err
		}
	}

	slog.Info("cache entries removed")

	return nil
}

// Removes cache entries matching a reference.
func removeReference(ctx context.Context, c *cache.Cache, refStr string) error {
	opts, err := reference.NewIdentifierOptions(internal.DefaultRegistryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}
	ref, err := reference.Parse(refStr, resource.TypeWidget, opts)
	if err != nil {
		return crex.UserError("invalid reference", err.Error()).
			Fallback("Use the format 'namespace/resource version'.").
			Err()
	}

	if ref.IsVersionBased() {
		return removeVersion(ctx, c, ref)
	}

	return removeAllVersions(ctx, c, ref)
}

// Removes a specific version from the cache.
func removeVersion(ctx context.Context, c *cache.Cache, ref *reference.Reference) error {
	_, err := c.Get(ctx, ref.Namespace(), ref.Name(), ref.Version().String())
	if errors.Is(err, cache.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return c.Delete(ctx, ref.Namespace(), ref.Name(), ref.Version().String())
}

// Removes all versions of a resource from the cache.
func removeAllVersions(ctx context.Context, c *cache.Cache, ref *reference.Reference) error {
	entries, err := c.List(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Namespace == ref.Namespace() && entry.Resource == ref.Name() {
			if err := c.Delete(ctx, entry.Namespace, entry.Resource, entry.String); err != nil {
				return err
			}
		}
	}

	return nil
}
