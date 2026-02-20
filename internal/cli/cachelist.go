package cli

import (
	"context"
	"fmt"

	"github.com/cruciblehq/crux/internal/cache"
)

// Represents the 'crux cache list' command.
type CacheListCmd struct {
	Namespace string `short:"n" help:"Filter by namespace."`
}

// Executes the cache list command.
func (c *CacheListCmd) Run(ctx context.Context) error {
	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return err
	}
	defer localCache.Close()

	entries, err := localCache.List(ctx)
	if err != nil {
		return err
	}

	if c.Namespace != "" {
		filtered := entries[:0]
		for _, entry := range entries {
			if entry.Namespace == c.Namespace {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}

	for _, entry := range entries {
		fmt.Printf("%s/%s %s\n", entry.Namespace, entry.Resource, entry.String)
	}

	return nil
}
