// Package archive provides functions for creating and extracting zstd-compressed
// tar archives with security-focused design.
//
// Archives are compressed using zstd. Only regular files and directories are
// supported; symlinks and special files (devices, sockets, named pipes) ar
// rejected with [ErrUnsupportedFileType].
//
// Path validation is enforced during extraction to prevent vulnerabilities:
//   - [filepath.Localize] validates and normalizes archive paths
//   - [filepath.IsLocal] ensures paths don't escape via absolute paths or ".."
//   - Symlinks are rejected to prevent symlink attacks
//   - Archive extraction fails atomically - partial extraction is cleaned up on error
//
// Example:
//
//	// Create an archive
//	err := archive.Create("mydir", "output.tar.zst")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Extract the archive
//	err = archive.Extract("output.tar.zst", "extracted")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Extract from an io.Reader
//	file, _ := os.Open("output.tar.zst")
//	defer file.Close()
//	err = archive.ExtractFromReader(file, "extracted")
//	if err != nil {
//		log.Fatal(err)
//	}
package archive
