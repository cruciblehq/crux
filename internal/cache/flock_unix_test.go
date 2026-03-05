//go:build !windows

package cache

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestLockExclusive(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "test.lock")

	// Acquire the lock on a first file descriptor.
	f1, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f1.Close()

	if err := lockFile(f1); err != nil {
		t.Fatal(err)
	}

	// A non-blocking attempt on a second descriptor should fail with EWOULDBLOCK.
	f2, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	err = syscall.Flock(int(f2.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == nil {
		t.Fatal("expected second lock to fail while first is held")
	}

	// Release the first lock and verify the second can now succeed.
	unlockFile(f1)

	if err := lockFile(f2); err != nil {
		t.Fatal("expected lock to succeed after unlock: ", err)
	}
	unlockFile(f2)
}
