//go:build !windows

package files

import (
	"context"
	"os"
	"syscall"
	"time"
)

// Interval between lock acquisition retries.
const retryInterval = 100 * time.Millisecond

// Acquires an exclusive file lock on f.
//
// Blocks until the lock is available or ctx is cancelled. Uses a non-blocking
// flock call in a retry loop so the operation honours context deadlines and
// cancellation without holding the OS scheduler. Returns an error if the lock
// cannot be acquired or ctx is cancelled.
func LockWithContext(ctx context.Context, f *os.File) error {
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
		case <-time.After(retryInterval):
		}
	}
}

// Acquires an exclusive file lock on f.
//
// Blocks until the lock is available. Returns an error if it cannot be acquired.
func Lock(f *os.File) error {
	return LockWithContext(context.Background(), f)
}

// Releases the file lock on f.
func Unlock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
