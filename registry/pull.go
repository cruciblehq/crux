package registry

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/reference"
)

// Holds the output of a successful [Source.Pull] call.
//
// Digest and Size always reflect the archive content regardless of whether
// it was newly downloaded or already present in the cache. Extracted is the
// path to the extracted archive on the local filesystem and remains valid
// until the cache entry is evicted.
type PullResult struct {
	Namespace string // Namespace name.
	Resource  string // Resource name.
	Version   string // Version string.
	Digest    string // Content digest.
	Size      int64  // Archive size in bytes.
	Extracted string // Local directory containing the extracted archive.
}

// Pulls a resource from the registry and extracts it locally.
//
// Opens the local cache, resolves the version through the registry (applying
// any channel or version constraint in ref), then returns a cached result if
// the stored digest matches, or downloads and stores the archive otherwise.
// The archive is extracted before returning; Extracted in the result points
// to the extraction root.
func pull(ctx context.Context, ref *reference.Reference) (*PullResult, error) {
	localCache, err := cache.Open()
	if err != nil {
		return nil, crex.Wrap(ErrCacheOperation, err)
	}
	defer localCache.Close()

	registryURL := ref.Registry()
	client := NewClient(registryURL, nil)

	ver, err := ResolveVersion(ctx, client, ref)
	if err != nil {
		return nil, crex.Wrap(ErrResolveVersion, err)
	}

	if ver.Digest == nil || *ver.Digest == "" {
		return nil, ErrNoArchive
	}

	var result *PullResult
	if cached, ok := checkCache(localCache, ref, ver, *ver.Digest); ok {
		result = cached
	} else {
		result, err = downloadAndCache(ctx, client, localCache, ref, ver, *ver.Digest)
		if err != nil {
			return nil, err
		}
	}

	result.Extracted, err = localCache.Extract(result.Namespace, result.Resource, result.Version)
	if err != nil {
		return nil, crex.Wrap(ErrCacheOperation, err)
	}

	return result, nil
}

// Returns a cached result when the entry exists and its digest matches expectedDigest.
//
// When the entry is present but has a stale or absent digest, the entry is
// removed from the cache on a best-effort basis and (nil, false) is returned.
func checkCache(c *cache.Cache, ref *reference.Reference, ver *Version, expectedDigest string) (*PullResult, bool) {
	entry, err := c.Get(ref.Namespace(), ref.Name(), ver.String)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			slog.Debug("resource cache miss", "namespace", ref.Namespace(), "resource", ref.Name(), "version", ver.String)
		} else {
			slog.Error("resource cache lookup failed", "error", err)
		}
		return nil, false
	}

	// Best effort to remove stale entry
	if entry.Digest == nil || *entry.Digest != expectedDigest {
		if err := c.Delete(ref.Namespace(), ref.Name(), ver.String); err != nil {
			slog.Error("failed to remove stale resource cache entry", "error", err)
		}
		return nil, false
	}

	return &PullResult{
		Namespace: ref.Namespace(),
		Resource:  ref.Name(),
		Version:   ver.String,
		Digest:    *entry.Digest,
		Size:      *entry.Size,
	}, true
}

// Downloads the archive from the registry and stores it in the cache.
//
// Verifies the stored archive against expectedDigest before returning;
// returns [ErrCacheOperation] if the digest does not match, indicating
// possible data corruption in transit.
func downloadAndCache(ctx context.Context, client *Client, c *cache.Cache, ref *reference.Reference, ver *Version, expectedDigest string) (*PullResult, error) {
	archiveReader, err := client.DownloadArchive(ctx, ref.Namespace(), ref.Name(), ver.String)
	if err != nil {
		return nil, crex.Wrap(ErrDownload, err)
	}
	defer archiveReader.Close()

	entry, err := c.PutWithDigest(ref.Namespace(), ref.Name(), ver.String, expectedDigest, archiveReader)
	if err != nil {
		return nil, crex.Wrap(ErrCacheOperation, err)
	}

	return &PullResult{
		Namespace: ref.Namespace(),
		Resource:  ref.Name(),
		Version:   ver.String,
		Digest:    *entry.Digest,
		Size:      *entry.Size,
	}, nil
}
