package provider

import "errors"

var (

	// Backend errors.

	ErrUnknownProvider = errors.New("unknown provider")

	// Client errors.

	ErrConnectionFailed  = errors.New("cruxd connection failed")
	ErrConnectionRefused = errors.New("cruxd connection refused")
	ErrRequestFailed     = errors.New("cruxd request failed")
	ErrNotRunning        = errors.New("cruxd is not running")
)
