package cli

// Manages the local resource cache.
type CacheCmd struct {
	List   *CacheListCmd   `cmd:"" help:"List all cached entries"`
	Clear  *CacheClearCmd  `cmd:"" help:"Clear all cached entries"`
	Remove *CacheRemoveCmd `cmd:"" help:"Remove specific cached entries"`
}
