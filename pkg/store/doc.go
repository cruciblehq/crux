// Package store provides local caching and remote fetching of Crucible resources.
//
// The store manages a local SQLite cache of resources fetched from the Crucible
// registry. It supports fetching namespace and resource metadata, downloading
// archives by version or channel, and caching results for offline use.
//
// The [Cache] provides persistent storage for registry metadata and downloaded
// archives. It stores namespace info, resource info (including versions and
// channels), and archive locations on disk.
//
//	cache, err := store.OpenCache()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close()
//
//	info, etag, err := cache.GetResource("crucible", "starter")
//
// The [Remote] interface communicates with the Crucible registry API. Use
// [NewRemote] to create a client for the official registry.
//
//	remote := store.NewRemote(store.RemoteOptions{
//	    UserAgent: "myapp/1.0",
//	})
//
//	resp, err := remote.Fetch(ctx, &store.FetchRequest{
//	    Identifier: ref.Identifier(),
//	    Version:    ref.Version(),
//	})
//
// Both the cache and remote support ETags for conditional requests. When
// fetching from the remote, include the cached ETag to receive a 304 Not
// Modified response if the content hasn't changed.
//
//	info, etag, err := cache.GetNamespace("crucible")
//	resp, err := remote.Namespace(ctx, &store.NamespaceRequest{
//	    Namespace: "crucible",
//	    ETag:      etag,
//	})
//	if resp.Info != nil {
//	    cache.PutNamespace(resp.Info, resp.ETag)
//	}
//
// The registry returns structured errors with content type [MediaTypeError].
// Use [errors.As] to check for registry errors:
//
//	var regErr *store.RegistryError
//	if errors.As(err, &regErr) {
//	    fmt.Printf("registry error: %s (%s)\n", regErr.Message, regErr.Code)
//	}
//
// The Crucible registry exposes resources at the following endpoints:
//
//	GET /v1/{namespace}                -> [MediaTypeNamespace]
//	GET /v1/{namespace}/{name}         -> [MediaTypeResource]
//	GET /v1/{namespace}/{name}/{ref}   -> [MediaTypeArchive]
//
// The {ref} parameter accepts either a semantic version (e.g., 1.2.0) or a
// channel prefixed with colon (e.g., :stable). Channels resolve to the latest
// version in that channel and return the Content-Version header with the
// resolved version.
//
// Note: Channel resolution is not atomic, which means that between calls to
// GET /v1/{namespace}/{name} and GET /v1/{namespace}/{name}/{ref}, the channel
// may have been updated to point to a different version. Channels are considered
// unreliable and are prohibited for production use.
package store
