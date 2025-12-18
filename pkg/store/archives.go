package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cruciblehq/crux/pkg/reference"
)

// Describes a locally extracted archive.
//
// Archives are downloaded from GET /v1/{namespace}/{name}/{ref} and extracted
// to disk. The Path field points to the extracted directory. Unlike other
// cached data, archives have no corresponding store type since they represent
// local filesystem state rather than registry metadata.
type Archive struct {
	Namespace string            // Parent namespace.
	Name      string            // Parent resource name.
	Version   reference.Version // Semantic version.
	Digest    reference.Digest  // Content digest for integrity verification.
	Path      string            // Filesystem path to extracted directory.
	ETag      string            // ETag for conditional requests.
}

// Returns cached archive metadata, or ErrNotFound.
//
// The Path field points to the extracted directory. The caller is responsible
// for verifying that the extracted directory at Path exists. If it doesn't,
// the caller should remove the cache entry and re-download the archive.
func (c *Cache) GetArchive(namespace, name string, version reference.Version) (*Archive, error) {
	var digestStr, path, etag string

	err := c.db.QueryRow(sqlArchivesGet, namespace, name, version.String()).Scan(&digestStr, &path, &etag)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("archive %s/%s@%s: %w", namespace, name, version.String(), ErrNotFound)
	}
	if err != nil {
		return nil, err
	}

	digest, err := reference.ParseDigest(digestStr)
	if err != nil {
		return nil, err
	}

	return &Archive{
		Namespace: namespace,
		Name:      name,
		Version:   version,
		Digest:    *digest,
		Path:      path,
		ETag:      etag,
	}, nil
}

// Records a downloaded archive.
//
// The caller is responsible for downloading and extracting the archive to disk
// before calling this method. The Path field must point to the extracted
// directory.
func (c *Cache) PutArchive(a *Archive) error {
	now := time.Now().Unix()

	_, err := c.db.Exec(sqlArchivesInsert,
		a.Namespace,
		a.Name,
		a.Version.String(),
		a.Digest.String(),
		a.Path,
		a.ETag,
		now,
		now,
	)
	return err
}

// Removes an archive record from the cache.
//
// This only removes the database entry. The caller is responsible for removing
// the extracted directory from disk.
func (c *Cache) DeleteArchive(namespace, name string, version reference.Version) error {
	_, err := c.db.Exec(sqlArchivesDelete, namespace, name, version.String())
	return err
}
