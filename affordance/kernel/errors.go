package kernel

import "errors"

var (
	ErrInvalidGrant = errors.New("invalid kernel grant")
	ErrInvalidSpec  = errors.New("invalid kernel spec")
)
