//go:build windows

package files

import (
	"context"
	"os"
)

// Acquires an exclusive file lock on f.
//
// File locking is not supported on Windows; always returns [ErrNotSupported].
func LockWithContext(_ context.Context, _ *os.File) error {
	return ErrNotSupported
}

// Acquires an exclusive file lock on f.
//
// File locking is not supported on Windows; always returns [ErrNotSupported].
func Lock(_ *os.File) error {
	return ErrNotSupported
}

// Releases the file lock on f.
//
// File locking is not supported on Windows; always returns [ErrNotSupported].
func Unlock(_ *os.File) error {
	return ErrNotSupported
}
