package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/registry"
)

const (

	// The lock filename within the cache directory.
	lockFilename = "cache.lock"

	// The data subdirectory within the cache directory.
	dataDir = "data"

	// Metadata filename within each version directory.
	metaFilename = "meta.json"

	// Archive filename within each version directory.
	archiveFilename = "archive.tar.zst"
)

// Provides thread-safe and process-safe access to the local cache.
//
// Stores cached artifacts as files organized into <namespace>/<resource>/<version>
// directories. Operations are protected by both an in-process mutex and a file
// lock. The cache must be closed when no longer needed to release the file lock.
type Cache struct {
	root string       // Cache root directory
	lock *os.File     // File lock handle
	mu   sync.RWMutex // Mutex for in-process synchronization
}

// Opens the local cache, creating it if necessary.
//
// A file lock is acquired to ensure exclusive write access across processes.
// The caller must call Close when done with the cache.
func Open(ctx context.Context, _ any) (*Cache, error) {
	return OpenAt(ctx, paths.Store())
}

// Opens a cache at the specified directory.
func OpenAt(_ context.Context, root string) (*Cache, error) {
	if err := os.MkdirAll(root, paths.DefaultDirMode); err != nil {
		return nil, err
	}

	lockPath := filepath.Join(root, lockFilename)
	lf, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, paths.DefaultFileMode)
	if err != nil {
		return nil, err
	}

	if err := lockFile(lf); err != nil {
		lf.Close()
		return nil, err
	}

	return &Cache{root: root, lock: lf}, nil
}

// Releases the file lock.
func (c *Cache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lock != nil {
		unlockFile(c.lock)
		err := c.lock.Close()
		c.lock = nil
		return err
	}
	return nil
}

// Checks whether an entry exists in the cache.
func (c *Cache) Has(_ context.Context, namespace, resource, version string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, err := os.Stat(c.metaPath(namespace, resource, version))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Retrieves version metadata from the cache.
//
// Returns ErrNotFound if the entry doesn't exist.
func (c *Cache) Get(_ context.Context, namespace, resource, version string) (*registry.Version, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.readMeta(namespace, resource, version)
}

// Returns a reader for a cached archive.
//
// The caller is responsible for closing the returned reader. Returns
// ErrNotFound if the entry doesn't exist or has no archive.
func (c *Cache) OpenArchive(_ context.Context, namespace, resource, version string) (io.ReadCloser, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	archivePath := c.archivePath(namespace, resource, version)
	f, err := os.Open(archivePath)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return f, err
}

// Stores an entry in the cache.
//
// Creates the directory structure and writes the archive and metadata. If an
// entry already exists for the same namespace/resource/version, it is replaced.
func (c *Cache) Put(_ context.Context, namespace, resource, version string, archive io.Reader) (*registry.Version, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.writeEntry(namespace, resource, version, archive)
}

// Stores an entry in the cache, verifying the digest.
//
// Like Put, but verifies that the archive's computed digest matches the
// expected digest. Returns ErrDigestMismatch if they don't match.
func (c *Cache) PutWithDigest(ctx context.Context, namespace, resource, version, expectedDigest string, archive io.Reader) (*registry.Version, error) {
	ver, err := c.Put(ctx, namespace, resource, version, archive)
	if err != nil {
		return nil, err
	}

	if ver.Digest != nil && *ver.Digest != expectedDigest {
		// Remove the bad entry
		os.RemoveAll(c.versionDir(namespace, resource, version))
		return nil, ErrDigestMismatch
	}

	return ver, nil
}

// Removes an entry from the cache.
//
// Returns nil if the entry doesn't exist (idempotent).
func (c *Cache) Delete(_ context.Context, namespace, resource, version string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir := c.versionDir(namespace, resource, version)
	err := os.RemoveAll(dir)
	if os.IsNotExist(err) {
		return nil
	}

	// Clean up empty parent directories.
	c.pruneEmpty(c.resourceDir(namespace, resource))
	c.pruneEmpty(c.namespaceDir(namespace))
	return err
}

// Returns all versions across all namespaces and resources.
func (c *Cache) List(_ context.Context) ([]*registry.Version, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var versions []*registry.Version
	root := filepath.Join(c.root, dataDir)

	namespaces, err := listSubdirs(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, ns := range namespaces {
		resources, _ := listSubdirs(filepath.Join(root, ns))
		for _, res := range resources {
			versionDirs, _ := listSubdirs(filepath.Join(root, ns, res))
			for _, ver := range versionDirs {
				v, err := c.readMeta(ns, res, ver)
				if err != nil {
					continue
				}
				versions = append(versions, v)
			}
		}
	}

	return versions, nil
}

// Removes all entries from the cache.
func (c *Cache) Clear(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	root := filepath.Join(c.root, dataDir)
	if err := os.RemoveAll(root); err != nil {
		return err
	}
	return nil
}

// Returns the data/<namespace> directory path.
func (c *Cache) namespaceDir(namespace string) string {
	return filepath.Join(c.root, dataDir, namespace)
}

// Returns the data/<namespace>/<resource> directory path.
func (c *Cache) resourceDir(namespace, resource string) string {
	return filepath.Join(c.root, dataDir, namespace, resource)
}

// Returns the data/<namespace>/<resource>/<version> directory path.
func (c *Cache) versionDir(namespace, resource, version string) string {
	return filepath.Join(c.root, dataDir, namespace, resource, version)
}

// Returns the path to the metadata file for a version.
func (c *Cache) metaPath(namespace, resource, version string) string {
	return filepath.Join(c.versionDir(namespace, resource, version), metaFilename)
}

// Returns the path to the archive file for a version.
func (c *Cache) archivePath(namespace, resource, version string) string {
	return filepath.Join(c.versionDir(namespace, resource, version), archiveFilename)
}

// Reads and parses the metadata file for a version.
func (c *Cache) readMeta(namespace, resource, version string) (*registry.Version, error) {
	data, err := os.ReadFile(c.metaPath(namespace, resource, version))
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var ver registry.Version
	if err := json.Unmarshal(data, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Writes an archive and its metadata to the version directory. Returns the
// populated Version with digest and size.
func (c *Cache) writeEntry(namespace, resource, version string, archive io.Reader) (*registry.Version, error) {
	dir := c.versionDir(namespace, resource, version)
	if err := os.MkdirAll(dir, paths.DefaultDirMode); err != nil {
		return nil, err
	}

	archivePath := c.archivePath(namespace, resource, version)

	// Write archive to a temp file first, computing digest as we go.
	tmpFile, err := os.CreateTemp(dir, ".archive-*.tmp")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up on failure.

	h := sha256.New()
	w := io.MultiWriter(tmpFile, h)

	size, err := io.Copy(w, archive)
	if err != nil {
		tmpFile.Close()
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	// Atomic rename.
	if err := os.Rename(tmpPath, archivePath); err != nil {
		return nil, err
	}

	digest := fmt.Sprintf("sha256:%s", hex.EncodeToString(h.Sum(nil)))
	now := time.Now().Unix()
	archiveStr := archiveFilename

	ver := &registry.Version{
		Namespace: namespace,
		Resource:  resource,
		String:    version,
		Archive:   &archiveStr,
		Size:      &size,
		Digest:    &digest,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Write metadata.
	meta, err := json.Marshal(ver)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(c.metaPath(namespace, resource, version), meta, paths.DefaultFileMode); err != nil {
		return nil, err
	}

	return ver, nil
}

// Removes a directory if it is empty.
func (c *Cache) pruneEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}
	os.Remove(dir)
}

// Lists immediate subdirectory names.
func listSubdirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
