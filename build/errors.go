package build

import "errors"

var (
	ErrBuild               = errors.New("build failed")
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidPath         = errors.New("invalid path")
	ErrInvalidFromFormat   = errors.New("invalid from format")
)
