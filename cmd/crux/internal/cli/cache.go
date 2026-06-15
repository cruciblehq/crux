package cli

// Manages the local resource cache.
type CacheCmd struct {
	List   *CacheListCmd   `cmd:"" aliases:"ls" help:"List all cached entries."`
	Clear  *CacheClearCmd  `cmd:"" help:"Clear all cached entries."`
	Remove *CacheRemoveCmd `cmd:"" aliases:"rm" help:"Remove specific cached entries."`
}
