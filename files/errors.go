package files

import "errors"

var ErrNotSupported = errors.New("file locking is not supported on this platform")
