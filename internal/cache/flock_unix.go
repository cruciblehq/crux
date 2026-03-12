//go:build !windows

package cache

import (
	"context"
	"os"
	"syscall"
	"time"
)

const (
	// Interval between non-blocking flock retries.
	lockRetryInterval = 100 * time.Millisecond
)

// Acquires an exclusive file lock, respecting context cancellation.
//
// Uses non-blocking flock with a retry loop so the caller can be interrupted
// by context cancellation or timeout. Falls back to a context-less blocking
// flock when ctx has no deadline.
func lockFile(f *os.File) error {
	return lockFileContext(context.Background(), f)
}

// Acquires an exclusive file lock, respecting context cancellation.
func lockFileContext(ctx context.Context, f *os.File) error {
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if err != syscall.EWOULDBLOCK {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(lockRetryInterval):
		}
	}
}

// Releases a file lock.
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
