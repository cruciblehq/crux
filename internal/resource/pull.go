package resource

import (
	"context"
	"errors"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/cache"
	"github.com/cruciblehq/crux/internal/registry"
	"github.com/cruciblehq/spec/reference"
	specregistry "github.com/cruciblehq/spec/registry"
)

// Holds the output of a successful [Pull] call.
//
// When Cached is true the archive already existed locally with the correct
// digest, so no download was performed. Digest and Size always reflect the
// archive content regardless of whether it was downloaded or cached.
type PullResult struct {
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
// downloaded from the registry and stored in the cache. Supports both
// version-based and channel-based references. Channels are resolved to their
// current version before downloading.
func pull(ctx context.Context, ref *reference.Reference) (*PullResult, error) {
	localCache, err := cache.Open()
	if err != nil {
		return nil, crex.Wrap(ErrCacheOperation, err)
	}
	defer localCache.Close()

	registryURL := ref.Registry()
	client := registry.NewClient(registryURL, nil)

	ver, err := registry.ResolveVersion(ctx, client, ref)
	if err != nil {
		return nil, handleResolveError(err)
	}

	if ver.Digest == nil || *ver.Digest == "" {
		return nil, crex.UserError("no archive", "version exists but has no uploaded archive").
			Fallback("The package may not have been fully pushed.").
			Err()
	}

	if result, ok := checkCache(localCache, ref, ver, *ver.Digest); ok {
		return result, nil
	}

	return downloadAndCache(ctx, client, localCache, ref, ver, *ver.Digest)
}

// Converts resolution errors to user-friendly errors.
func handleResolveError(err error) error {
	if errors.Is(err, registry.ErrNoVersions) {
		return crex.UserError("no versions found", "resource has no pushed versions").
			Fallback("Ensure the resource exists and has at least one version.").
			Err()
	}
	if errors.Is(err, registry.ErrNoMatchingVersion) {
		return crex.UserError("no matching version", "no version satisfies the constraint").
			Fallback("Check the version constraint and available versions.").
			Err()
	}
	if errors.Is(err, registry.ErrTypeMismatch) {
		return crex.UserError("type mismatch", "the resource type does not match what was requested").
			Fallback("Ensure the resource type matches what you requested.").
			Cause(err).
			Err()
	}

	var regErr *specregistry.Error
	if errors.As(err, &regErr) && regErr.Code == specregistry.ErrorCodeNotFound {
		return crex.UserError("not found", regErr.Message).
			Fallback("Check the resource name and try again.").
			Err()
	}

	return crex.UserError("failed to resolve version", "could not determine the target version").
		Fallback("Check your network connection and registry URL.").
		Cause(err).
		Err()
}

// Returns a cached result if the entry exists with matching digest.
func checkCache(c *cache.Cache, ref *reference.Reference, ver *specregistry.Version, expectedDigest string) (*PullResult, bool) {
	entry, err := c.Get(ref.Namespace(), ref.Name(), ver.String)
	if err != nil {
		return nil, false
	}

	if entry.Digest == nil || *entry.Digest != expectedDigest {
		c.Delete(ref.Namespace(), ref.Name(), ver.String)
		return nil, false
	}

	return &PullResult{
		Namespace: ref.Namespace(),
		Resource:  ref.Name(),
		Version:   ver.String,
		Digest:    *entry.Digest,
		Size:      *entry.Size,
		Cached:    true,
	}, true
}

// Downloads the archive and stores it in the cache.
func downloadAndCache(ctx context.Context, client *registry.Client, c *cache.Cache, ref *reference.Reference, ver *specregistry.Version, expectedDigest string) (*PullResult, error) {
	archiveReader, err := client.DownloadArchive(ctx, ref.Namespace(), ref.Name(), ver.String)
	if err != nil {
		return nil, crex.UserError("failed to download archive", "could not retrieve the archive from the registry").
			Fallback("Check your network connection and try again.").
			Cause(err).
			Err()
	}
	defer archiveReader.Close()

	entry, err := c.PutWithDigest(ref.Namespace(), ref.Name(), ver.String, expectedDigest, archiveReader)
	if err != nil {
		if errors.Is(err, cache.ErrDigestMismatch) {
			return nil, crex.UserError("digest mismatch", "downloaded archive doesn't match expected digest").
				Fallback("The archive may have been corrupted in transit. Try again.").
				Err()
		}
		return nil, crex.Wrap(ErrCacheOperation, err)
	}

	return &PullResult{
		Namespace: ref.Namespace(),
		Resource:  ref.Name(),
		Version:   ver.String,
		Digest:    *entry.Digest,
		Size:      *entry.Size,
		Cached:    false,
	}, nil
}
