package cli

import "errors"

var (
	ErrUnexpectedState = errors.New("unexpected runtime state")
	ErrImport          = errors.New("import failed")
)
