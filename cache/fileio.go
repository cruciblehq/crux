package cache

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/utils-go/archive"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Opens a file, returning ErrNotFound if it doesn't exist.
func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return f, err
}

// Extracts a zstd-compressed tar archive into dir atomically.
//
// The archive is extracted into a temporary sibling directory, then renamed
// into place.
func extractDirAtomic(r io.Reader, dir string) error {
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, file.DefaultDirMode); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp(parent, ".extracting-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) // No-op after successful rename.

	if err := archive.ExtractFromReader(r, tmpDir, archive.Zstd); err != nil {
		return err
	}

	return os.Rename(tmpDir, dir)
}

// Reads and parses a version metadata file.
func readMeta(metaPath string) (*Version, error) {
	const recoveryCacheClear = "Run 'crux cache clear' and try again."

	data, err := os.ReadFile(metaPath)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, crex.SystemError("cannot read cache", "failed to read the cache metadata file").
			Recovery(recoveryCacheClear).
			Cause(err).
			Err()
	}

	var ver Version
	if err := json.Unmarshal(data, &ver); err != nil {
		return nil, crex.SystemError("cannot read cache", "the cache metadata file is corrupt").
			Recovery(recoveryCacheClear).
			Cause(err).
			Err()
	}
	return &ver, nil
}

// Builds version metadata and writes it as JSON to metPath.
func writeMeta(metPath, namespace, resource, version, digest string, size int64) (*Version, error) {
	now := time.Now().Unix()
	archiveStr := archiveFilename

	ver := &Version{
		Namespace: namespace,
		Resource:  resource,
		String:    version,
		Archive:   &archiveStr,
		Size:      &size,
		Digest:    &digest,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(ver)
	if err != nil {
		return nil, crex.SystemError("cannot write to cache", "failed to encode the cache metadata").
			Recovery("Run 'crux cache clear' and try again.").
			Cause(err).
			Err()
	}
	if err := os.WriteFile(metPath, data, file.DefaultFileMode); err != nil {
		return nil, crex.SystemError("cannot write to cache", "failed to write the cache metadata file").
			Recovery("Free up disk space, then try again.").
			Cause(err).
			Err()
	}
	return ver, nil
}
