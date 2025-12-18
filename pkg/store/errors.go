package store

import "errors"

var (
	ErrCacheRequired     = errors.New("cache required")
	ErrRemoteRequired    = errors.New("remote required")
	ErrNotFound          = errors.New("not found")
	ErrNoMatchingVersion = errors.New("no matching version")
	ErrDigestMismatch    = errors.New("digest mismatch")
	ErrFetchFailed       = errors.New("fetch failed")
	ErrExtractFailed     = errors.New("extraction failed")
)
