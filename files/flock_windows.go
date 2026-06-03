//go:build windows

package files

import (
	"context"
	"os"
)

func LockWithContext(_ context.Context, _ *os.File) error {
	return ErrNotSupported
}

func Lock(_ *os.File) error {
	return ErrNotSupported
}

func Unlock(_ *os.File) error {
	return ErrNotSupported
}
