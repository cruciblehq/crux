//go:build windows

package fio

import (
	"context"
	"os"
)

// Not supported on Windows.
func LockWithContext(_ context.Context, _ *os.File) error {
	return ErrNotSupported
}

// Not supported on Windows.
func Lock(_ *os.File) error {
	return ErrNotSupported
}

// Not supported on Windows.
func Unlock(_ *os.File) error {
	return ErrNotSupported
}
