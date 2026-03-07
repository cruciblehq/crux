package local

import "errors"

var (

	// Runtime environment errors.

	ErrUnsupportedPlatform   = errors.New("unsupported platform")
	ErrRuntimeNotCreated     = errors.New("runtime has not been created")
	ErrRuntimeAlreadyRunning = errors.New("runtime is already running")
	ErrRuntimeNotRunning     = errors.New("runtime is not running")
	ErrRuntimeStart          = errors.New("failed to start runtime")
	ErrRuntimeStop           = errors.New("failed to stop runtime")
	ErrRuntimeDestroy        = errors.New("failed to destroy runtime")
	ErrRuntimeExec           = errors.New("failed to execute command in runtime")
	ErrRuntimeConfig         = errors.New("failed to generate runtime configuration")

	// Lima, image, and cruxd installation errors.

	ErrLimaDownload  = errors.New("failed to download lima")
	ErrImageDownload = errors.New("failed to download machine image")
	ErrCruxdInstall  = errors.New("failed to install cruxd")
)
