//go:build !windows

package cache

import (
	"os"
	"syscall"
)

// Acquires an exclusive file lock (blocks until available).
func lockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// Releases a file lock.
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
