package cli

import "github.com/cruciblehq/crux/crex"

var (
	ErrUnexpectedState = crex.New("unexpected runtime state")
	ErrImport          = crex.New("import failed")
)
