package widget

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrInvalidPath = crex.New("invalid build path")
	ErrBuild       = crex.New("build failed")
)
