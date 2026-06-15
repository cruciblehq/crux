package widget

import "github.com/cruciblehq/crux/crex"

var (
	ErrInvalidPath = crex.New("invalid build path")
	ErrBuild       = crex.New("build failed")
)
