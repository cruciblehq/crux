package files

import "github.com/cruciblehq/crux/crex"

var (
	ErrNotSupported  = crex.New("file locking is not supported on this platform")
	ErrInvalidPath   = crex.New("invalid path")
	ErrUnsafeTempDir = crex.New("unsafe temporary directory")
)
