package local

import "github.com/cruciblehq/crux/crex"

var (
	ErrUnsupportedPlatform    = crex.New("unsupported platform")
	ErrHostNotCreated         = crex.New("host has not been created")
	ErrHostAlreadyProvisioned = crex.New("host is already provisioned")
	ErrHostNotRunning         = crex.New("host is not running")
	ErrHostCreate             = crex.New("failed to create host")
	ErrHostStart              = crex.New("failed to start host")
	ErrHostStop               = crex.New("failed to stop host")
	ErrHostDestroy            = crex.New("failed to destroy host")
	ErrMachineImageMissing    = crex.New("machine image not in cache")
	ErrImageUpload            = crex.New("failed to upload image")
	ErrHostExec               = crex.New("failed to execute command in host")
	ErrHostConfig             = crex.New("failed to generate host configuration")
	ErrLimaDownload           = crex.New("failed to download lima")
	ErrLimaCtl                = crex.New("limactl command failed")
)
