package cli

import "errors"

var (
	ErrFileSystem      = errors.New("filesystem operation failed")
	ErrUnexpectedState = errors.New("unexpected runtime state")
)
