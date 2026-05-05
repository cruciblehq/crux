package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/internal/cache"
)

// Represents the 'crux cache clear' command.
type CacheClearCmd struct{}

// Executes the cache clear command.
func (c *CacheClearCmd) Run(_ context.Context) error {
	slog.Info("clearing cache...")

	localCache, err := cache.Open()
	if err != nil {
		return err
	}
	defer localCache.Close()

	if err := localCache.Clear(); err != nil {
		return err
	}

	slog.Info("cache cleared")

	return nil
}
