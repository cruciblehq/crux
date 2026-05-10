package source

import "errors"

var (
	ErrMissingOption       = errors.New("missing required option")
	ErrMissingRegistry     = errors.New("default registry is required")
	ErrMissingNamespace    = errors.New("default namespace is required")
	ErrCacheOperation      = errors.New("cache operation failed")
	ErrFileSystemOperation = errors.New("file system operation failed")
	ErrNoArchive           = errors.New("version has no uploaded archive")
	ErrResolveVersion      = errors.New("failed to resolve version")
	ErrDownload            = errors.New("failed to download archive")
	ErrInvalidResource     = errors.New("invalid resource identifier")
	ErrNamespaceNotFound   = errors.New("namespace not found")
	ErrVersionExists       = errors.New("version already exists")
	ErrRegistryOperation   = errors.New("registry operation failed")
)
