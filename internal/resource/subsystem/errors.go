package subsystem

import "errors"

var (
	ErrUnsupportedDomain = errors.New("unsupported domain")
	ErrSeccompExpression = errors.New("invalid seccomp expression")
	ErrSeccompArgFilter  = errors.New("invalid seccomp arg filter")
	ErrMACExpression     = errors.New("invalid mac expression")
	ErrSandboxExpression = errors.New("invalid sandbox expression")
)
