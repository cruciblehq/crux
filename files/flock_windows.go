//go:build windows

package files

import (
	"context"
	"os"

	"github.com/cruciblehq/crux/crex"
)

// System error returned by every file-locking operation.
var errNotSupported = crex.SystemError("file locking is unavailable", "file locking is not supported on Windows").
	Cause(ErrNotSupported).
	Err()

// Acquires an exclusive file lock on f.
//
// File locking is not supported on Windows; always returns [ErrNotSupported].
func LockWithContext(_ context.Context, _ *os.File) error {
	return errNotSupported
}

// Acquires an exclusive file lock on f.
//
// File locking is not supported on Windows; always returns [ErrNotSupported].
func Lock(_ *os.File) error {
	return errNotSupported
}

// Releases the file lock on f.
//
// File locking is not supported on Windows; always returns [ErrNotSupported].
func Unlock(_ *os.File) error {
	return errNotSupported
}
