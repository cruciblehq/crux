package archive

import "errors"

var (

	// Returned when archive creation fails.
	ErrCreateFailed = errors.New("archive creation failed")

	// Returned when archive extraction fails.
	ErrExtractFailed = errors.New("extraction failed")

	// Returned when a path is invalid or attempts directory traversal.
	ErrInvalidPath = errors.New("invalid path")

	// Returned when a symlink is encountered. Symlinks are forbidden in
	// resource archives because they could point to arbitrary locations
	// on the filesystem when extracted to a different directory.
	ErrSymlink = errors.New("symlinks are not allowed")

	// Returned when an unsupported file type is encountered. Only regular
	// files and directories are allowed in resource archives.
	ErrUnsupportedFileType = errors.New("unsupported file type")

	// Returned when the destination directory already exists.
	ErrDestinationExists = errors.New("destination already exists")
)
