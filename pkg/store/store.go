package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cruciblehq/crux/pkg/archive"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/crux/pkg/reference"
)

// Manages local caching and retrieval of Crucible resources.
//
// Coordinates between the remote registry and local cache. When resolving a
// reference, it checks the cache first and uses ETags for conditional requests
// to avoid unnecessary downloads. Safe for concurrent use.
type Store struct {
	cache    *Cache // Local cache (database).
	remote   Remote // Remote registry (HTTP).
	basePath string // Base path for extracted archives.
	mu       sync.RWMutex
}

// Creates a new Store with the given cache and remote.
//
// The cache stores registry metadata and tracks extracted archives. The remote
// fetches resource info and archives from the registry. Both are required.
func New(cache *Cache, remote Remote) (*Store, error) {
	return newWithPath(cache, remote, paths.Store())
}

// Creates a new Store with a custom base path (for testing).
func newWithPath(cache *Cache, remote Remote, basePath string) (*Store, error) {
	if cache == nil {
		return nil, ErrCacheRequired
	}
	if remote == nil {
		return nil, ErrRemoteRequired
	}

	return &Store{
		cache:    cache,
		remote:   remote,
		basePath: basePath,
	}, nil
}

// Resolves a reference to a local filesystem path.
//
// The resolution process:
//  1. Check for a cached archive matching the resolved version
//  2. If not cached, fetch resource info from the remote
//  3. Select the best matching version based on the reference
//  4. Download and extract the archive if not already cached
//
// For channel-based references (e.g., "namespace/name:stable"), the version
// associated with that channel is selected.
//
// For version-constrained references (e.g., "namespace/name@^1.0.0"), the
// highest matching version is selected.
//
// Returns the absolute path to the extracted resource directory.
func (s *Store) Resolve(ctx context.Context, ref *reference.Reference) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Fetch resource info from cache (if available) or remote.
	info, err := s.getResourceInfo(ctx, ref.Identifier)
	if err != nil {
		return "", err
	}

	// Select the version to use.
	version, err := s.selectVersion(ref, info)
	if err != nil {
		return "", err
	}

	// Is the archive cached?
	arc, err := s.cache.GetArchive(ref.Identifier.Namespace(), ref.Identifier.Name(), *version)
	if err == nil {
		if _, statErr := os.Stat(arc.Path); statErr == nil {
			return arc.Path, nil
		}

		// Path missing, remove stale cache entry.
		s.cache.DeleteArchive(ref.Identifier.Namespace(), ref.Identifier.Name(), *version)
	}

	// Not cached; fetch, extract, and cache.
	return s.fetchAndExtract(ctx, ref.Identifier, version, info)
}

// Fetches resource info, using the cache for conditional requests.
//
// If the cached ETag matches the server's version, the cached info is returned.
// Otherwise, the fresh info from the server is cached and returned.
func (s *Store) getResourceInfo(ctx context.Context, id reference.Identifier) (*ResourceInfo, error) {
	cached, etag, _ := s.cache.GetResource(id.Namespace(), id.Name())

	req := &ResourceRequest{
		Identifier: id,
		ETag:       etag,
	}

	resp, err := s.remote.Resource(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchFailed, err)
	}

	// 304 Not Modified; use cached.
	if resp.Info == nil {
		return cached, nil
	}

	if len(resp.Info.Versions) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id.String())
	}

	// Cache the response.
	s.cache.PutResource(id.Namespace(), resp.Info, resp.ETag)

	return resp.Info, nil
}

// Fetches and extracts an archive to the local filesystem.
func (s *Store) fetchAndExtract(ctx context.Context, id reference.Identifier, v *reference.Version, info *ResourceInfo) (string, error) {
	digest, err := findDigest(v, info)
	if err != nil {
		return "", err
	}

	req := &FetchRequest{
		Identifier: id,
		Version:    *v,
	}

	resp, err := s.remote.Fetch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrFetchFailed, err)
	}

	// Verify the archive integrity.
	if err := verifyDigest(resp.Data, digest); err != nil {
		return "", err
	}

	extractPath := s.versionPath(id, v)

	if err := archive.ExtractReader(bytes.NewReader(resp.Data), extractPath); err != nil {
		return "", fmt.Errorf("%w: %w", ErrExtractFailed, err)
	}

	// Record the archive in the cache.
	s.cache.PutArchive(&Archive{
		Namespace: id.Namespace(),
		Name:      id.Name(),
		Version:   *v,
		Digest:    *digest,
		Path:      extractPath,
		ETag:      resp.ETag,
	})

	return extractPath, nil
}

// Finds the digest for a specific version.
func findDigest(v *reference.Version, info *ResourceInfo) (*reference.Digest, error) {
	for it := range info.Versions {
		if info.Versions[it].Version.String() == v.String() {
			return &info.Versions[it].Digest, nil
		}
	}
	return nil, fmt.Errorf("%w: version %s", ErrNoMatchingVersion, v.String())
}

// Verifies the archive data matches the expected digest.
func verifyDigest(data []byte, expected *reference.Digest) error {
	hash := sha256.Sum256(data)
	actual := "sha256:" + hex.EncodeToString(hash[:])

	if actual != expected.String() {
		return fmt.Errorf("%w: expected %s, got %s", ErrDigestMismatch, expected.String(), actual)
	}

	return nil
}

// Selects the best matching version for a reference.
//
// For channel-based references, returns the version associated with that
// channel. For version-constrained references, returns the highest version
// that satisfies the constraint.
func (s *Store) selectVersion(ref *reference.Reference, info *ResourceInfo) (*reference.Version, error) {
	if ref.IsChannelBased() {
		return s.selectByChannel(ref, info.Channels)
	}

	return s.selectByConstraint(ref, info.Versions)
}

// Finds the version associated with a channel.
func (s *Store) selectByChannel(ref *reference.Reference, channels []ChannelInfo) (*reference.Version, error) {
	channel := *ref.Channel()

	for it := range channels {
		if channels[it].Channel == channel {
			return &channels[it].Version, nil
		}
	}

	return nil, fmt.Errorf("%w: channel %s", ErrNoMatchingVersion, channel)
}

// Finds the highest version matching a constraint.
//
// When comparing prerelease versions with different identifiers (e.g., alpha
// vs beta), the comparison is skipped since they are not directly comparable.
func (s *Store) selectByConstraint(ref *reference.Reference, versions []VersionInfo) (*reference.Version, error) {
	constraint := ref.Version()

	var latest *reference.Version
	for it := range versions {
		v := &versions[it].Version
		matches, err := constraint.MatchesVersion(v)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrNoMatchingVersion, err)
		}
		if matches {
			if latest == nil {
				latest = v
				continue
			}
			cmp, ok := v.Compare(latest)
			if ok && cmp > 0 {
				latest = v
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("%w: %s", ErrNoMatchingVersion, constraint.String())
	}

	return latest, nil
}

// Returns the extraction path for a specific version.
func (s *Store) versionPath(id reference.Identifier, v *reference.Version) string {
	return filepath.Join(s.basePath, id.Namespace(), id.Name(), v.String())
}
