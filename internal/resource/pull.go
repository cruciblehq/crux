package resource

import (
	"context"
	"errors"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/cache"
	"github.com/cruciblehq/crux/internal/reference"
	"github.com/cruciblehq/crux/internal/registry"
)

// Holds the output of a successful [Source.Pull] call.
//
// Digest and Size always reflect the archive content regardless of whether
// it was freshly downloaded or already present in the cache.
type PullResult struct {
	Namespace string // Namespace name.
	Resource  string // Resource name.
	Version   string // Version string.
	Digest    string // Content digest.
	Size      int64  // Archive size in bytes.
	Dir       string // Local directory containing the extracted archive.
}

// Pulls a resource from the registry and extracts it locally.
//
// If the resource is already in the cache with the correct digest, no download
// occurs. Otherwise, the archive is downloaded from the registry and stored in
// the cache. The archive is then extracted and Dir is set to the extraction
// directory. Supports both version-based and channel-based references.
// Channels are resolved to their current version before downloading.
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

	var result *PullResult
	if cached, ok := checkCache(localCache, ref, ver, *ver.Digest); ok {
		result = cached
	} else {
		result, err = downloadAndCache(ctx, client, localCache, ref, ver, *ver.Digest)
		if err != nil {
			return nil, err
		}
	}

	result.Dir, err = localCache.Extract(result.Namespace, result.Resource, result.Version)
	if err != nil {
		return nil, crex.Wrap(ErrCacheOperation, err)
	}

	return result, nil
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

	var regErr *registry.Error
	if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
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
func checkCache(c *cache.Cache, ref *reference.Reference, ver *registry.Version, expectedDigest string) (*PullResult, bool) {
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
	}, true
}

// Downloads the archive and stores it in the cache.
func downloadAndCache(ctx context.Context, client *registry.Client, c *cache.Cache, ref *reference.Reference, ver *registry.Version, expectedDigest string) (*PullResult, error) {
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
	}, nil
}
