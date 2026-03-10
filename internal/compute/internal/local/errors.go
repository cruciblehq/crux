package local

import "errors"

var (

	// Host environment errors.

	ErrUnsupportedPlatform = errors.New("unsupported platform")
	ErrHostNotCreated      = errors.New("host has not been created")
	ErrHostAlreadyRunning  = errors.New("host is already running")
	ErrHostNotRunning      = errors.New("host is not running")
	ErrHostStart           = errors.New("failed to start host")
	ErrHostStop            = errors.New("failed to stop host")
	ErrHostDestroy         = errors.New("failed to destroy host")
	ErrHostExec            = errors.New("failed to execute command in host")
	ErrHostConfig          = errors.New("failed to generate host configuration")

	// Lima and cruxd installation errors.

	ErrLimaDownload = errors.New("failed to download lima")
	ErrCruxdInstall = errors.New("failed to install cruxd")
)
