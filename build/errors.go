package build

import "errors"

var (
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidPath         = errors.New("invalid path")
)
