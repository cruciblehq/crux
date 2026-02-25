package resource

import "errors"

var (
	ErrBuild               = errors.New("build failed")
	ErrRunner              = errors.New("runner failed")
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidStructure    = errors.New("invalid resource structure")
	ErrInvalidPath         = errors.New("invalid path")
	ErrCacheOperation      = errors.New("cache operation failed")
	ErrUnsupported         = errors.New("unsupported operation")
)
