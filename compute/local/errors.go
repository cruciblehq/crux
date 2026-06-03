package local

import "errors"

var (
	ErrUnsupportedPlatform    = errors.New("unsupported platform")
	ErrHostNotCreated         = errors.New("host has not been created")
	ErrHostAlreadyProvisioned = errors.New("host is already provisioned")
	ErrHostNotRunning         = errors.New("host is not running")
	ErrHostCreate             = errors.New("failed to create host")
	ErrHostStart              = errors.New("failed to start host")
	ErrHostStop               = errors.New("failed to stop host")
	ErrHostDestroy            = errors.New("failed to destroy host")
	ErrMachineImageMissing    = errors.New("machine image not in cache")
	ErrImageUpload            = errors.New("failed to upload image")
	ErrHostExec               = errors.New("failed to execute command in host")
	ErrHostConfig             = errors.New("failed to generate host configuration")
	ErrLimaDownload           = errors.New("failed to download lima")
	ErrLimaCtl                = errors.New("limactl command failed")
)
