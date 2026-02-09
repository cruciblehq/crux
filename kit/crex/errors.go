package crex

import "errors"

var (
	ErrEmptyDescription = errors.New("description is empty")
	ErrEmptyReason      = errors.New("reason is empty")
)
