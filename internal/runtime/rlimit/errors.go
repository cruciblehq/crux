package rlimit

import "errors"

var (
	ErrInvalidGrant = errors.New("invalid rlimit grant")
	ErrConflict     = errors.New("rlimit conflict")
)
