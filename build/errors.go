package build

import "errors"

var (
	ErrCopy                = errors.New("copy failed")
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidPath         = errors.New("invalid path")
	ErrInvalidFromFormat   = errors.New("invalid from format")
)
