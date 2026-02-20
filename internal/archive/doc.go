// Package archive provides functions for creating and extracting compressed tar
// archives.
//
// Supported formats are Zstandard (.tar.zst) and Gzip (.tar.gz, .tgz). The
// compression format is detected automatically from the file extension by
// [Create] and [Extract], or supplied explicitly to [ExtractFromReader]. Only
// regular files and directories are archivable; symlinks and special files
// (devices, sockets, named pipes) are rejected with [ErrUnsupportedFileType]
// during creation. Extraction additionally supports symlinks and hard links,
// both validated against directory escape. Path traversal attacks and absolute
// paths are detected and rejected with [ErrInvalidPath].
//
// Created archives preserve source file permissions with special bits (setuid,
// setgid, sticky) stripped. Extracted entries preserve permissions from the
// archive headers with the same masking. Intermediate directories that have no
// archive entry are created with [os.ModePerm], subject to the process umask.
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
//	err := archive.Extract("output.tar.gz", "extracted")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Extracting from an [io.Reader]:
//
//	file, _ := os.Open("output.tar.zst")
//	defer file.Close()
//	err := archive.ExtractFromReader(file, "extracted", archive.Zstd)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Reading a single file from a tar stream:
//
//	tr := tar.NewReader(r)
//	data, err := archive.Find(tr, "crucible.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if data == nil {
//		log.Fatal("file not found in archive")
//	}
package archive
