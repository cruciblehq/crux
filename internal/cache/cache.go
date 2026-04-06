package cache

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/registry"
)

const (

	// The lock filename within the cache directory.
	lockFilename = "cache.lock"

	// The archives subdirectory within the cache directory.
	archivesDir = "archives"

	// The extracted subdirectory within the cache directory.
	extractedDir = "extracted"

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
func Open() (*Cache, error) {
	return OpenAt(paths.RegistryCacheDir())
}

// Opens a cache at the specified directory.
//
// A file lock is acquired to ensure exclusive write access across processes.
// The caller must call Close when done with the cache.
func OpenAt(root string) (*Cache, error) {
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

// Closes the cache and releases the file lock.
func (c *Cache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lock != nil {
		err := errors.Join(
			unlockFile(c.lock),
			c.lock.Close(),
		)
		c.lock = nil
		return err
	}
	return nil
}

// Checks whether an entry exists in the cache.
func (c *Cache) Has(namespace, resource, version string) (bool, error) {
	meta, err := c.metaPath(namespace, resource, version)
	if err != nil {
		return false, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return pathExists(meta)
}

// Retrieves version metadata from the cache.
func (c *Cache) Get(namespace, resource, version string) (*registry.Version, error) {
	meta, err := c.metaPath(namespace, resource, version)
	if err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return readMeta(meta)
}

// Returns a reader for a cached archive.
//
// The caller is responsible for closing the returned reader. The cache's read
// lock is held until the reader is closed, preventing concurrent writes from
// modifying the archive during the read.
func (c *Cache) OpenArchive(namespace, resource, version string) (io.ReadCloser, error) {
	path, err := c.archivePath(namespace, resource, version)
	if err != nil {
		return nil, err
	}
	c.mu.RLock()

	f, err := openFile(path)
	if err != nil {
		c.mu.RUnlock()
		return nil, err
	}
	return &lockedReadCloser{file: f, unlock: c.mu.RUnlock}, nil
}

// Checks whether an entry has been extracted.
func (c *Cache) HasExtracted(namespace, resource, version string) (bool, error) {
	dir, err := c.extractedVersionDir(namespace, resource, version)
	if err != nil {
		return false, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return pathExists(dir)
}

// Returns the path to the extracted contents of a cached archive.
//
// If the archive has already been extracted, the existing path is returned.
// Otherwise, the archive is extracted atomically into the extracted tree.
// The archive format is assumed to be Zstandard-compressed tar (tar.zst).
func (c *Cache) Extract(namespace, resource, version string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.extract(namespace, resource, version)
}

// Stores an entry in the cache.
//
// Creates the directory structure and writes the archive and metadata. If an
// entry already exists for the same namespace/resource/version, it is removed
// first (including any extracted contents).
func (c *Cache) Put(namespace, resource, version string, archive io.Reader) (*registry.Version, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.put(namespace, resource, version, archive)
}

// Stores an entry in the cache, verifying the digest.
//
// Like Put, but verifies that the archive's computed digest matches the
// expected digest. Returns ErrDigestMismatch if they don't match.
func (c *Cache) PutWithDigest(namespace, resource, version, expectedDigest string, archive io.Reader) (*registry.Version, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ver, err := c.put(namespace, resource, version, archive)
	if err != nil {
		return nil, err
	}

	if *ver.Digest != expectedDigest {
		if err := c.removeVersion(namespace, resource, version); err != nil {
			slog.Error("failed to clean up after digest mismatch", "error", err)
		}
		return nil, ErrDigestMismatch
	}

	return ver, nil
}

// Removes an entry from the cache, including any extracted contents.
//
// Returns nil if the entry doesn't exist (idempotent).
func (c *Cache) Delete(namespace, resource, version string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.removeVersion(namespace, resource, version)
}

// Returns all versions across all namespaces and resources.
func (c *Cache) List() ([]*registry.Version, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list()
}

// Removes all entries from the cache, including extracted contents.
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.clear()
}

// Returns the archives root directory path.
func (c *Cache) archivesRoot() string {
	return filepath.Join(c.root, archivesDir)
}

// Returns the extracted root directory path.
func (c *Cache) extractedRoot() string {
	return filepath.Join(c.root, extractedDir)
}

// Returns the archives/<namespace>/<resource>/<version> directory path.
func (c *Cache) versionDir(namespace, resource, version string) (string, error) {
	return safeJoin(c.archivesRoot(), namespace, resource, version)
}

// Returns the extracted/<namespace>/<resource>/<version> directory path.
func (c *Cache) extractedVersionDir(namespace, resource, version string) (string, error) {
	return safeJoin(c.extractedRoot(), namespace, resource, version)
}

// Returns the path to the metadata file for a version.
func (c *Cache) metaPath(namespace, resource, version string) (string, error) {
	dir, err := c.versionDir(namespace, resource, version)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, metaFilename), nil
}

// Returns the path to the archive file for a version.
func (c *Cache) archivePath(namespace, resource, version string) (string, error) {
	dir, err := c.versionDir(namespace, resource, version)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, archiveFilename), nil
}

// Opens the archive file for a version. Returns ErrNotFound if the entry
// doesn't exist or has no archive.
func (c *Cache) openArchiveFile(namespace, resource, version string) (*os.File, error) {
	path, err := c.archivePath(namespace, resource, version)
	if err != nil {
		return nil, err
	}
	return openFile(path)
}

// Extracts a cached archive into the extracted tree, returning the directory path.
func (c *Cache) extract(namespace, resource, version string) (string, error) {
	dir, err := c.extractedVersionDir(namespace, resource, version)
	if err != nil {
		return "", err
	}

	// Already extracted.
	exists, err := pathExists(dir)
	if err != nil {
		return "", err
	}
	if exists {
		return dir, nil
	}

	// Open the cached archive.
	f, err := c.openArchiveFile(namespace, resource, version)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := extractDirAtomic(f, dir); err != nil {
		return "", err
	}

	return dir, nil
}

// Removes any existing entry and writes a new one.
func (c *Cache) put(namespace, resource, version string, archive io.Reader) (*registry.Version, error) {
	if err := c.removeVersion(namespace, resource, version); err != nil {
		return nil, err
	}
	return c.writeEntry(namespace, resource, version, archive)
}

// Walks the archives tree and returns all stored versions.
func (c *Cache) list() ([]*registry.Version, error) {
	root := c.archivesRoot()

	namespaces, err := listSubdirs(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var versions []*registry.Version
	for _, ns := range namespaces {
		versions = append(versions, listNamespace(root, ns)...)
	}
	return versions, nil
}

// Returns all versions within a single namespace directory.
func listNamespace(root, ns string) []*registry.Version {
	resources, err := listSubdirs(filepath.Join(root, ns))
	if err != nil {
		slog.Error("failed to list resources", "namespace", ns, "error", err)
		return nil
	}

	var versions []*registry.Version
	for _, res := range resources {
		versionDirs, err := listSubdirs(filepath.Join(root, ns, res))
		if err != nil {
			slog.Error("failed to list versions", "namespace", ns, "resource", res, "error", err)
			continue
		}
		for _, ver := range versionDirs {
			metPath := filepath.Join(root, ns, res, ver, metaFilename)
			v, err := readMeta(metPath)
			if err != nil {
				slog.Error("failed to read metadata", "namespace", ns, "resource", res, "version", ver, "error", err)
				continue
			}
			versions = append(versions, v)
		}
	}
	return versions
}

// Removes both the archives and extracted trees.
func (c *Cache) clear() error {
	return errors.Join(
		os.RemoveAll(c.archivesRoot()),
		os.RemoveAll(c.extractedRoot()),
	)
}

// Creates the directory structure, writes the archive, and stores metadata.
func (c *Cache) writeEntry(namespace, resource, version string, r io.Reader) (*registry.Version, error) {
	dir, err := c.versionDir(namespace, resource, version)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, paths.DefaultDirMode); err != nil {
		return nil, err
	}

	archPath := filepath.Join(dir, archiveFilename)
	digest, size, err := writeFileAtomic(r, dir, archPath)
	if err != nil {
		return nil, err
	}

	metPath := filepath.Join(dir, metaFilename)
	return writeMeta(metPath, namespace, resource, version, digest, size)
}

// Removes a version's archive and extracted contents, pruning empty parent
// directories afterward.
func (c *Cache) removeVersion(namespace, resource, version string) error {
	vDir, err := c.versionDir(namespace, resource, version)
	if err != nil {
		return err
	}
	eDir, err := c.extractedVersionDir(namespace, resource, version)
	if err != nil {
		return err
	}

	err = errors.Join(
		os.RemoveAll(vDir),
		os.RemoveAll(eDir),
	)

	// Best-effort parent directory cleanup.
	for _, dir := range []string{
		filepath.Dir(vDir),
		filepath.Dir(filepath.Dir(vDir)),
		filepath.Dir(eDir),
		filepath.Dir(filepath.Dir(eDir)),
	} {
		if pErr := pruneEmpty(dir); pErr != nil {
			slog.Error("failed to prune empty directory", "dir", dir, "error", pErr)
		}
	}

	return err
}

// Joins base with safe path components. Each component must be a non-empty
// directory name, not "." or "..", and containing no path separators.
func safeJoin(base string, components ...string) (string, error) {
	for _, c := range components {
		if err := validatePathComponent(c); err != nil {
			return "", err
		}
	}
	return filepath.Join(append([]string{base}, components...)...), nil
}

// Validates a single path component.
func validatePathComponent(s string) error {
	if s == "" {
		return errors.New("empty path component")
	}
	if s == "." || s == ".." {
		return crex.Wrapf(ErrInvalidPath, "%q", s)
	}
	if strings.ContainsAny(s, "/\\") {
		return crex.Wrapf(ErrInvalidPath, "%q contains separator", s)
	}
	return nil
}
