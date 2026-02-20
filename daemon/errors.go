package daemon

import "errors"

var (
	ErrConnection = errors.New("daemon connection failed")
	ErrRequest    = errors.New("daemon request failed")
	ErrNotRunning = errors.New("daemon is not running")
)
