// Package archive provides functions for creating and extracting zstd-compressed
// tar archives.
//
// Archives are compressed using Zstandard (zstd). Only regular files and
// directories are supported; symlinks and special files (devices, sockets,
// named pipes) are rejected with [ErrUnsupportedFileType]. Path traversal
// attacks and absolute paths are detected and rejected with [ErrInvalidPath].
//
// Creating an archive from a directory:
//
//	err := archive.Create("mydir", "output.tar.zst")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Extracting an archive to a new directory:
//
//	err := archive.Extract("output.tar.zst", "extracted")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Extracting from an [io.Reader]:
//
//	file, _ := os.Open("output.tar.zst")
//	defer file.Close()
//	err := archive.ExtractFromReader(file, "extracted")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Reading a single file from a tar stream:
//
//	tr := tar.NewReader(r)
//	data, err := archive.FindInTar(tr, "crucible.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if data == nil {
//		log.Fatal("file not found in archive")
//	}
package archive
