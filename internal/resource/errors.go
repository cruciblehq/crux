package resource

import "errors"

var (
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResource     = errors.New("invalid resource format")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidStructure    = errors.New("invalid resource structure")
	ErrCacheOperation      = errors.New("cache operation failed")
	ErrVersionNotFound     = errors.New("version not found")
	ErrNoArchive           = errors.New("version has no archive")
)
