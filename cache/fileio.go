package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Opens a file, returning [ErrNotFound] if it doesn't exist.
func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return f, err
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
	if err := codec.Decode(data, &ver, codec.JSON); err != nil {
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

	data, err := codec.Encode(ver, codec.JSON)
	if err != nil {
		return nil, crex.SystemError("cannot write to cache", "failed to encode the cache metadata").
			Recovery("Run 'crux cache clear' and try again.").
			Cause(err).
			Err()
	}
	if err := file.WriteFile(metPath, data, file.DefaultFileMode); err != nil {
		return nil, crex.SystemError("cannot write to cache", "failed to write the cache metadata file").
			Recovery("Free up disk space, then try again.").
			Cause(err).
			Err()
	}
	return ver, nil
}

// Streams r to path atomically and returns its SHA-256 digest and size.
//
// The digest is hex encoded with a "sha256:" prefix, the form recorded in the
// entry metadata.
func writeArchive(r io.Reader, path string) (string, int64, error) {
	a, err := file.CreateAtomic(filepath.Dir(path), file.DefaultFileMode)
	if err != nil {
		return "", 0, err
	}
	defer a.Abort()

	h := sha256.New()
	size, err := io.Copy(io.MultiWriter(a, h), r)
	if err != nil {
		return "", 0, err
	}
	if err := a.Commit(path); err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(h.Sum(nil))), size, nil
}
