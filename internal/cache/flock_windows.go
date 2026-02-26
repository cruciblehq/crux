//go:build windows

package cache

import "os"

// Not supported on Windows.
func lockFile(_ *os.File) error {
	panic("file locking is not supported on Windows")
}

// Not supported on Windows.
func unlockFile(_ *os.File) {
	panic("file locking is not supported on Windows")
}
