package push

import "errors"

var (
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidResource     = errors.New("invalid resource format")
)
