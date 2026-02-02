package cache

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/protocol/pkg/registry"
)

const (
	// The database filename within the cache directory.
	databaseFilename = "cache.db"

	// The lock filename within the cache directory.
	lockFilename = "cache.lock"

	// The archives subdirectory within the cache directory.
	archivesDir = "archives"
)

// Provides thread-safe and process-safe access to the local registry.
//
// Operations are protected by both an in-process mutex and a file lock. The
// cache must be closed when no longer needed to release the file lock.
type Cache struct {
	root     string                // Cache root directory
	registry *registry.SQLRegistry // Underlying registry
	db       *sql.DB               // Database connection (for cleanup)
	lockFile *os.File              // File lock handle
	mu       sync.RWMutex          // In-process mutex
}

// Opens the local cache, creating it if necessary.
//
// A file lock is acquired to ensure exclusive write access across processes.
// The caller must call Close when done with the cache.
func Open(ctx context.Context, _ any) (*Cache, error) {
	return OpenAt(ctx, paths.Store())
}

// Opens a cache at the specified directory.
func OpenAt(ctx context.Context, root string) (*Cache, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(root, paths.DefaultDirMode); err != nil {
		return nil, err
	}

	// Acquire file lock
	lockPath := filepath.Join(root, lockFilename)
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, paths.DefaultFileMode)
	if err != nil {
		return nil, err
	}

	// Acquire exclusive lock (blocks until available)
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		lockFile.Close()
		return nil, err
	}

	// Open SQLite database
	dbPath := filepath.Join(root, databaseFilename)
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		return nil, err
	}

	// Create registry with archives directory
	archivesPath := filepath.Join(root, archivesDir)
	reg, err := registry.NewSQLRegistry(ctx, db, archivesPath, nil)
	if err != nil {
		db.Close()
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		return nil, err
	}

	return &Cache{
		root:     root,
		registry: reg,
		db:       db,
		lockFile: lockFile,
	}, nil
}

// Releases the file lock and closes the database connection.
func (c *Cache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	if c.db != nil {
		if err := c.db.Close(); err != nil {
			errs = append(errs, err)
		}
		c.db = nil
	}

	if c.lockFile != nil {
		syscall.Flock(int(c.lockFile.Fd()), syscall.LOCK_UN)
		if err := c.lockFile.Close(); err != nil {
			errs = append(errs, err)
		}
		c.lockFile = nil
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Checks whether an entry exists in the cache.
func (c *Cache) Has(ctx context.Context, namespace, resource, version string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, err := c.registry.ReadVersion(ctx, namespace, resource, version)
	if err != nil {
		var regErr *registry.Error
		if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Retrieves version metadata from the cache.
//
// Returns ErrNotFound if the entry doesn't exist.
func (c *Cache) Get(ctx context.Context, namespace, resource, version string) (*registry.Version, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ver, err := c.registry.ReadVersion(ctx, namespace, resource, version)
	if err != nil {
		var regErr *registry.Error
		if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ver, nil
}

// Stores an entry in the cache.
//
// Creates the namespace, resource, and version if they don't exist, then
// uploads the archive. If an entry already exists for the same
// namespace/resource/version, the archive is replaced.
func (c *Cache) Put(ctx context.Context, namespace, resource, version string, archive io.Reader) (*registry.Version, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureNamespace(ctx, namespace); err != nil {
		return nil, err
	}

	if err := c.ensureResource(ctx, namespace, resource); err != nil {
		return nil, err
	}

	if err := c.ensureVersion(ctx, namespace, resource, version); err != nil {
		return nil, err
	}

	return c.registry.UploadArchive(ctx, namespace, resource, version, archive)
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
		c.registry.DeleteVersion(ctx, namespace, resource, version)
		return nil, ErrDigestMismatch
	}

	return ver, nil
}

// Removes an entry from the cache.
//
// Returns nil if the entry doesn't exist (idempotent).
func (c *Cache) Delete(ctx context.Context, namespace, resource, version string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.registry.DeleteVersion(ctx, namespace, resource, version)
	if err != nil {
		var regErr *registry.Error
		if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
			return nil
		}
		return err
	}
	return nil
}

// Returns all versions across all namespaces and resources.
func (c *Cache) List(ctx context.Context) ([]*registry.Version, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var versions []*registry.Version

	nsList, err := c.registry.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	for _, ns := range nsList.Namespaces {
		resList, err := c.registry.ListResources(ctx, ns.Name)
		if err != nil {
			continue
		}

		for _, res := range resList.Resources {
			verList, err := c.registry.ListVersions(ctx, ns.Name, res.Name)
			if err != nil {
				continue
			}

			for _, verSum := range verList.Versions {
				ver, err := c.registry.ReadVersion(ctx, ns.Name, res.Name, verSum.String)
				if err != nil {
					continue
				}
				versions = append(versions, ver)
			}
		}
	}

	return versions, nil
}

// Removes all entries from the cache.
func (c *Cache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	nsList, err := c.registry.ListNamespaces(ctx)
	if err != nil {
		return err
	}

	for _, ns := range nsList.Namespaces {
		resList, err := c.registry.ListResources(ctx, ns.Name)
		if err != nil {
			continue
		}

		for _, res := range resList.Resources {
			verList, err := c.registry.ListVersions(ctx, ns.Name, res.Name)
			if err != nil {
				continue
			}

			for _, ver := range verList.Versions {
				c.registry.DeleteVersion(ctx, ns.Name, res.Name, ver.String)
			}

			c.registry.DeleteResource(ctx, ns.Name, res.Name)
		}

		c.registry.DeleteNamespace(ctx, ns.Name)
	}

	return nil
}

// Opens the archive file for reading.
//
// The caller is responsible for closing the returned reader.
func (c *Cache) OpenArchive(ctx context.Context, namespace, resource, version string) (io.ReadCloser, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.registry.DownloadArchive(ctx, namespace, resource, version)
}

// Ensures a namespace exists, creating it if necessary.
func (c *Cache) ensureNamespace(ctx context.Context, namespace string) error {
	_, err := c.registry.ReadNamespace(ctx, namespace)
	if err == nil {
		return nil
	}

	var regErr *registry.Error
	if !errors.As(err, &regErr) || regErr.Code != registry.ErrorCodeNotFound {
		return err
	}

	_, err = c.registry.CreateNamespace(ctx, registry.NamespaceInfo{
		Name: namespace,
	})
	if err != nil {
		var createErr *registry.Error
		if errors.As(err, &createErr) && createErr.Code == registry.ErrorCodeNamespaceExists {
			return nil // Race condition: another process created it
		}
		return err
	}
	return nil
}

// Ensures a resource exists, creating it if necessary.
func (c *Cache) ensureResource(ctx context.Context, namespace, resource string) error {
	_, err := c.registry.ReadResource(ctx, namespace, resource)
	if err == nil {
		return nil
	}

	var regErr *registry.Error
	if !errors.As(err, &regErr) || regErr.Code != registry.ErrorCodeNotFound {
		return err
	}

	_, err = c.registry.CreateResource(ctx, namespace, registry.ResourceInfo{
		Name: resource,
	})
	if err != nil {
		var createErr *registry.Error
		if errors.As(err, &createErr) && createErr.Code == registry.ErrorCodeResourceExists {
			return nil
		}
		return err
	}
	return nil
}

// Ensures a version exists, creating it if necessary.
func (c *Cache) ensureVersion(ctx context.Context, namespace, resource, version string) error {
	_, err := c.registry.ReadVersion(ctx, namespace, resource, version)
	if err == nil {
		return nil
	}

	var regErr *registry.Error
	if !errors.As(err, &regErr) || regErr.Code != registry.ErrorCodeNotFound {
		return err
	}

	_, err = c.registry.CreateVersion(ctx, namespace, resource, registry.VersionInfo{
		String: version,
	})
	if err != nil {
		var createErr *registry.Error
		if errors.As(err, &createErr) && createErr.Code == registry.ErrorCodeVersionExists {
			return nil
		}
		return err
	}
	return nil
}
