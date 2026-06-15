package cache

import "github.com/cruciblehq/crux/crex"

var (
	ErrNotFound       = crex.New("entry not found in cache")
	ErrDigestMismatch = crex.New("archive digest mismatch")
	ErrInvalidPath    = crex.New("invalid cache path")
)
