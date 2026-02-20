package pull

import "errors"

var (
	ErrCacheOperation  = errors.New("cache operation failed")
	ErrVersionNotFound = errors.New("version not found")
	ErrNoArchive       = errors.New("version has no archive")
)
