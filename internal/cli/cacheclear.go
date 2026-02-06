package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/cache"
)

// Represents the 'crux cache clear' command.
type CacheClearCmd struct{}

// Executes the cache clear command.
func (c *CacheClearCmd) Run(ctx context.Context) error {
	slog.Info("clearing cache...")

	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return err
	}
	defer localCache.Close()

	if err := localCache.Clear(ctx); err != nil {
		return err
	}

	slog.Info("cache cleared")

	return nil
}
