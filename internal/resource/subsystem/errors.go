package subsystem

import "errors"

var (
	ErrUnsupportedDomain = errors.New("unsupported domain")
	ErrGrantConflict     = errors.New("grant conflict")
	ErrSeccompExpression = errors.New("invalid seccomp expression")
	ErrSeccompArgFilter  = errors.New("invalid seccomp arg filter")
	ErrMACExpression     = errors.New("invalid mac expression")
	ErrGrantExpression   = errors.New("invalid grant expression")
)
