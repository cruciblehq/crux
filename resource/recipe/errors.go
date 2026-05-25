package recipe

import "errors"

var (
	ErrBuild               = errors.New("build failed")
	ErrFileSystemOperation = errors.New("file system operation failed")
)
