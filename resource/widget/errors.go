package widget

import "errors"

var (
	ErrInvalidPath = errors.New("invalid build path")
	ErrBuild       = errors.New("build failed")
)
