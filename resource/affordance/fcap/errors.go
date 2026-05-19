package fcap

import "errors"

var (
	ErrInvalidGrant = errors.New("invalid fcap grant")
	ErrConflict     = errors.New("fcap conflict")
)
