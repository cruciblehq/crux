package lima

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrUnsupportedPlatform       = crex.New("unsupported platform")
	ErrMachineNotCreated         = crex.New("machine has not been created")
	ErrMachineAlreadyProvisioned = crex.New("machine is already provisioned")
	ErrMachineNotRunning         = crex.New("machine is not running")
	ErrMachineCreate             = crex.New("failed to create machine")
	ErrMachineStart              = crex.New("failed to start machine")
	ErrMachineStop               = crex.New("failed to stop machine")
	ErrMachineDestroy            = crex.New("failed to destroy machine")
	ErrMachineImageMissing       = crex.New("machine image not in cache")
	ErrImageUpload               = crex.New("failed to upload image")
	ErrMachineExec               = crex.New("failed to execute command in machine")
	ErrMachineConfig             = crex.New("failed to generate machine configuration")
	ErrLimaDownload              = crex.New("failed to download lima")
	ErrLimaCtl                   = crex.New("limactl command failed")
)
