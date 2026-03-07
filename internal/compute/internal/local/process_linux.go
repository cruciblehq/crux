//go:build linux

package local

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Blocks until the process with the given PID exits.
//
// Uses pidfd_open to obtain a file descriptor for the process, then ppoll to
// wait for the descriptor to become readable. Returns immediately if the
// process does not exist (already exited).
func waitForProcessExit(pid int) {
	fd, _, errno := syscall.RawSyscall(unix.SYS_PIDFD_OPEN, uintptr(pid), 0, 0)
	if errno != 0 {
		return // ESRCH = already exited, other = can't watch
	}
	defer syscall.Close(int(fd))

	pfd := struct {
		fd      int32
		events  int16
		revents int16
	}{
		fd:     int32(fd),
		events: 1, // POLLIN
	}

	for {
		_, _, errno := syscall.Syscall6(unix.SYS_PPOLL, uintptr(unsafe.Pointer(&pfd)), 1, 0, 0, 0, 0)
		if errno == syscall.EINTR {
			continue
		}
		return // process exited or unrecoverable error
	}
}
