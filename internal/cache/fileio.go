package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/archive"
	"github.com/cruciblehq/spec/registry"
)

// Checks whether a path exists on the filesystem.
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Opens a file, returning ErrNotFound if it doesn't exist.
func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return f, err
}

// Extracts a zstd-compressed tar archive into dir atomically. The archive is
// extracted into a temporary sibling directory, then renamed into place.
func extractDirAtomic(r io.Reader, dir string) error {
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, paths.DefaultDirMode); err != nil {
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

// Writes r to destPath atomically via a temp file in tmpDir, computing the
// SHA-256 digest. Returns the digest string and the number of bytes written.
func writeFileAtomic(r io.Reader, tmpDir, destPath string) (string, int64, error) {
	tmpFile, err := os.CreateTemp(tmpDir, ".tmp-*")
	if err != nil {
		return "", 0, err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // No-op after successful rename.

	h := sha256.New()
	w := io.MultiWriter(tmpFile, h)

	size, err := io.Copy(w, r)
	if err != nil {
		tmpFile.Close()
		return "", 0, err
	}
	if err := tmpFile.Close(); err != nil {
		return "", 0, err
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return "", 0, err
	}
	if err := os.Chmod(destPath, paths.DefaultFileMode); err != nil {
		return "", 0, err
	}

	digest := fmt.Sprintf("sha256:%s", hex.EncodeToString(h.Sum(nil)))
	return digest, size, nil
}

// Reads and parses a version metadata file.
func readMeta(metaPath string) (*registry.Version, error) {
	data, err := os.ReadFile(metaPath)
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

// Builds version metadata and writes it as JSON to metPath.
func writeMeta(metPath, namespace, resource, version, digest string, size int64) (*registry.Version, error) {
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

	data, err := json.Marshal(ver)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(metPath, data, paths.DefaultFileMode); err != nil {
		return nil, err
	}
	return ver, nil
}

// Removes a directory if it is empty.
func pruneEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(entries) > 0 {
		return nil
	}
	err = os.Remove(dir)
	if os.IsNotExist(err) {
		return nil
	}
	return err
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
