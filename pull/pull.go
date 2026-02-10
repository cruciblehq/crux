package pull

import (
	"context"
	"errors"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/registry"
	"github.com/cruciblehq/crux/resource"
)

// Options for pulling a resource.
type Options struct {
	Registry         string        // Hub registry URL.
	Reference        string        // Resource reference (e.g., namespace/resource 1.0.0).
	Type             resource.Type // Resource type context for parsing.
	DefaultNamespace string        // Default namespace for resource identifiers.
}

// Result contains information about the pull operation.
type Result struct {
	Namespace string // Namespace name.
	Resource  string // Resource name.
	Version   string // Version string.
	Digest    string // Content digest.
	Size      int64  // Archive size in bytes.
	Cached    bool   // True if already in cache (no download).
}

// Downloads a resource from the registry and stores it in the local cache.
//
// If the resource is already in the cache with the correct digest, no download
// occurs and Cached is set to true in the result. Otherwise, the archive is
// downloaded from the registry and stored in the cache. Supports both version-
// based references (namespace/resource 1.0.0) and channel-based references
// (namespace/resource :stable). Channels are resolved to their current version
// before downloading.
func Pull(ctx context.Context, opts Options) (*Result, error) {
	refOpts, err := reference.NewIdentifierOptions(opts.Registry, opts.DefaultNamespace)
	if err != nil {
		return nil, err
	}
	ref, err := reference.Parse(opts.Reference, opts.Type, refOpts)
	if err != nil {
		return nil, crex.UserError("invalid reference", err.Error()).
			Fallback("Use the format 'namespace/resource version'.").
			Err()
	}

	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return nil, crex.Wrap(ErrCacheOperation, err)
	}
	defer localCache.Close()

	client := registry.NewClient(opts.Registry, nil)

	ver, err := registry.ResolveVersion(ctx, client, ref)
	if err != nil {
		return nil, handleResolveError(err)
	}

	if ver.Digest == nil || *ver.Digest == "" {
		return nil, crex.UserError("no archive", "version exists but has no uploaded archive").
			Fallback("The package may not have been fully published.").
			Err()
	}

	if result, ok := checkCache(ctx, localCache, ref, ver, *ver.Digest); ok {
		return result, nil
	}

	return downloadAndCache(ctx, client, localCache, ref, ver, *ver.Digest)
}

// Converts resolution errors to user-friendly errors.
func handleResolveError(err error) error {
	if errors.Is(err, registry.ErrNoVersions) {
		return crex.UserError("no versions found", "resource has no published versions").
			Fallback("Ensure the resource exists and has at least one version.").
			Err()
	}
	if errors.Is(err, registry.ErrNoMatchingVersion) {
		return crex.UserError("no matching version", "no version satisfies the constraint").
			Fallback("Check the version constraint and available versions.").
			Err()
	}
	if errors.Is(err, registry.ErrTypeMismatch) {
		return crex.UserError("type mismatch", err.Error()).
			Fallback("Ensure the resource type matches what you requested.").
			Err()
	}

	var regErr *registry.Error
	if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
		return crex.UserError("not found", regErr.Message).
			Fallback("Check the resource name and try again.").
			Err()
	}

	return crex.UserError("failed to resolve version", err.Error()).
		Fallback("Check your network connection and registry URL.").
		Err()
}

// Returns a cached result if the entry exists with matching digest.
func checkCache(ctx context.Context, c *cache.Cache, ref *reference.Reference, ver *registry.Version, expectedDigest string) (*Result, bool) {
	entry, err := c.Get(ctx, ref.Namespace(), ref.Name(), ver.String)
	if err != nil {
		return nil, false
	}

	if entry.Digest == nil || *entry.Digest != expectedDigest {
		c.Delete(ctx, ref.Namespace(), ref.Name(), ver.String)
		return nil, false
	}

	return &Result{
		Namespace: ref.Namespace(),
		Resource:  ref.Name(),
		Version:   ver.String,
		Digest:    *entry.Digest,
		Size:      *entry.Size,
		Cached:    true,
	}, true
}

// Downloads the archive and stores it in the cache.
func downloadAndCache(ctx context.Context, client *registry.Client, c *cache.Cache, ref *reference.Reference, ver *registry.Version, expectedDigest string) (*Result, error) {
	archiveReader, err := client.DownloadArchive(ctx, ref.Namespace(), ref.Name(), ver.String)
	if err != nil {
		return nil, crex.UserError("failed to download archive", err.Error()).
			Fallback("Check your network connection and try again.").
			Err()
	}
	defer archiveReader.Close()

	entry, err := c.PutWithDigest(ctx, ref.Namespace(), ref.Name(), ver.String, expectedDigest, archiveReader)
	if err != nil {
		if errors.Is(err, cache.ErrDigestMismatch) {
			return nil, crex.UserError("digest mismatch", "downloaded archive doesn't match expected digest").
				Fallback("The archive may have been corrupted in transit. Try again.").
				Err()
		}
		return nil, crex.Wrap(ErrCacheOperation, err)
	}

	return &Result{
		Namespace: ref.Namespace(),
		Resource:  ref.Name(),
		Version:   ver.String,
		Digest:    *entry.Digest,
		Size:      *entry.Size,
		Cached:    false,
	}, nil
}
