package seccomp

import "github.com/cruciblehq/crux/crex"

var (
	ErrInvalidGrant   = crex.New("invalid seccomp grant")
	ErrUnknownSyscall = crex.New("unknown syscall")
)
