package pack

import "errors"

var (
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidStructure    = errors.New("invalid resource structure")
)
