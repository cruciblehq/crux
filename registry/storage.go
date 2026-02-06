package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/kit/archive"
)

const (

	// Suffix for temporary upload files
	TemporaryUploadSuffix = ".upload.tmp"
)

// Returns the directory path for storing archives of a specific version.
//
// The path follows the pattern: {archiveRoot}/{namespace}/{resource}/{version}
func (r *SQLRegistry) archiveDirectoryPath(namespace, resource, version string) string {
	return filepath.Join(r.archiveRoot, namespace, resource, version)
}

// Returns the path for a temporary upload file in the archive directory.
//
// Temporary files use the .upload.tmp suffix during the upload process.
func (r *SQLRegistry) archiveTempPath(namespace, resource, version string) string {
	archiveDir := r.archiveDirectoryPath(namespace, resource, version)
	return filepath.Join(archiveDir, TemporaryUploadSuffix)
}

// Returns the final path for an archive file based on its digest.
//
// The path follows the pattern: {archiveDir}/{digest}.tar.zst
func (r *SQLRegistry) archiveFinalPath(namespace, resource, version, digest string) string {
	archiveDir := r.archiveDirectoryPath(namespace, resource, version)
	return filepath.Join(archiveDir, digest+archive.ArchiveFileExtension)
}

// Stores an archive file to disk and calculates its digest.
//
// Writes the archive data to a temporary file while calculating the SHA-256
// digest, then moves it to the final location named by the digest. Returns
// the digest, final file path, and size in bytes.
func (r *SQLRegistry) storeArchiveFile(namespace, resource, version string, archiveReader io.Reader) (digest, path string, size int64, err error) {

	// Create archive directory structure
	archiveDir := r.archiveDirectoryPath(namespace, resource, version)
	if err := os.MkdirAll(archiveDir, archive.DirMode); err != nil {
		return "", "", 0, err
	}

	// Create temporary file to store archive while calculating digest
	tempPath := r.archiveTempPath(namespace, resource, version)
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", "", 0, err
	}

	// Copy archive data while calculating SHA-256 digest and size
	hasher := sha256.New()
	writer := io.MultiWriter(tempFile, hasher)
	size, err = io.Copy(writer, archiveReader)
	if err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return "", "", 0, err
	}

	// Calculate digest from hash
	digest = fmt.Sprintf("sha256:%s", hex.EncodeToString(hasher.Sum(nil)))

	// Close temp file before rename (required on Windows)
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return "", "", 0, err
	}

	// Move temporary file to final location with digest-based name
	path = r.archiveFinalPath(namespace, resource, version, digest)
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", "", 0, err
	}

	return digest, path, size, nil
}
