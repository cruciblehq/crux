package cache

import "errors"

var (
	ErrNotFound       = errors.New("entry not found in cache")
	ErrDigestMismatch = errors.New("archive digest mismatch")
	ErrInvalidPath    = errors.New("invalid path component")
)
