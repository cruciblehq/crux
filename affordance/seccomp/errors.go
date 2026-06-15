package seccomp

import "errors"

var (
	ErrInvalidGrant   = errors.New("invalid seccomp grant")
	ErrUnknownSyscall = errors.New("unknown syscall")
)
