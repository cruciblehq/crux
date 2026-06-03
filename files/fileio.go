package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Removes dir if it contains no entries.
//
// If dir does not exist, returns nil. If dir is not empty, returns nil and does
// not remove it. Returns an error if dir exists but cannot be read or removed.
func RemoveDirIfEmpty(dir string) error {
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

// Lists the names of immediate subdirectories of dir.
//
// Regular files are ignored. The returned slice is nil when dir contains no
// subdirectories, and the names are in directory-read order (not sorted).
func ListSubdirs(dir string) ([]string, error) {
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

// Whether path exists.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Writes r to destPath atomically via a temporary file in tmpDir.
//
// The data is streamed through a SHA-256 hash during the write. On success the
// temporary file is renamed to destPath and the function returns the hex-encoded
// digest (prefixed with "sha256:") and the number of bytes written. No partial
// file is left behind on failure.
func WriteAtomic(r io.Reader, tmpDir, destPath string) (string, int64, error) {
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
	if err := os.Chmod(destPath, DefaultFileMode); err != nil {
		return "", 0, err
	}

	digest := fmt.Sprintf("sha256:%s", hex.EncodeToString(h.Sum(nil)))
	return digest, size, nil
}
