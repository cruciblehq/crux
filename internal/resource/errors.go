package resource

import "errors"

var (
	ErrBuild               = errors.New("build failed")
	ErrReadManifest        = errors.New("failed to read manifest")
	ErrResolveBuilder      = errors.New("failed to resolve builder")
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrInvalidPath         = errors.New("invalid path")
	ErrCacheOperation      = errors.New("cache operation failed")
	ErrUnsupported         = errors.New("unsupported operation")
)
