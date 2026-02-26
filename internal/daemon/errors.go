package daemon

import "errors"

var (
	ErrConnectionFailed  = errors.New("daemon connection failed")
	ErrConnectionRefused = errors.New("daemon connection refused")
	ErrRequestFailed     = errors.New("daemon request failed")
	ErrNotRunning        = errors.New("daemon is not running")
)
