package runtime

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupportedPlatform    = errors.New("unsupported platform")
	ErrVMNotCreated           = errors.New("vm has not been created")
	ErrVMAlreadyRunning       = errors.New("vm is already running")
	ErrVMNotRunning           = errors.New("vm is not running")
	ErrVMStart                = errors.New("failed to start vm")
	ErrVMStop                 = errors.New("failed to stop vm")
	ErrVMExec                 = errors.New("failed to execute command in vm")
	ErrVMConfig               = errors.New("failed to generate vm configuration")
	ErrHTTPStatus             = errors.New("unexpected HTTP status")
	ErrContainerd             = errors.New("failed to connect to containerd")
	ErrContainerdDownload     = errors.New("failed to download containerd")
	ErrContainerdNotFound     = errors.New("containerd binary not found in archive")
	ErrContainerdConfig       = errors.New("failed to generate containerd configuration")
	ErrContainerdStart        = errors.New("failed to start containerd")
	ErrContainerdStop         = errors.New("failed to stop containerd")
	ErrContainerdTimeout      = errors.New("timed out waiting for containerd to start")
	ErrContainerdExited       = errors.New("containerd exited unexpectedly")
	ErrVMSSH                  = errors.New("failed to establish ssh connection to vm")
	ErrSSHIdentityKey         = errors.New("invalid SSH identity key")
	ErrSSHPort                = errors.New("failed to query SSH port")
	ErrSSHPortNotAssigned     = errors.New("vm has no SSH port assigned")
	ErrLimaNotFound           = errors.New("lima binary not found")
	ErrLimaDownload           = errors.New("failed to download lima")
	ErrLimactlNotInArchive    = errors.New("limactl not found in archive")
	ErrVMNotAvailable         = errors.New("vm not available")
	ErrVMStatusCheck          = errors.New("unable to check vm status")
	ErrContainerdNotAvailable = errors.New("containerd not available")
	ErrContainerdStatusCheck  = errors.New("unable to check containerd status")
	ErrContainerdNotRunning   = errors.New("containerd is not running")
	ErrResourceRef            = errors.New("failed to parse resource ref")
	ErrImageFileOpen          = errors.New("failed to open image file")
	ErrImageImport            = errors.New("failed to import image into containerd")
	ErrImageDestroy           = errors.New("failed to destroy image")
	ErrContainerStart         = errors.New("failed to start container")
	ErrContainerStop          = errors.New("failed to stop container")
	ErrContainerDestroy       = errors.New("failed to destroy container")
	ErrContainerExec          = errors.New("failed to execute command in container")
	ErrContainerStatus        = errors.New("failed to query container status")
)

// A failed limactl command.
//
// The subcommand name, exit code, and raw output are stored separately so
// callers can present a clean message to users while preserving diagnostic
// detail for debug logging.
type commandError struct {
	subcommand string // The limactl subcommand that failed (e.g. "start").
	exitCode   int    // Process exit code.
	output     string // Combined stdout/stderr from the process.
}

// User-facing summary without raw CLI output.
func (e *commandError) Error() string {
	return fmt.Sprintf("limactl %s exited with code %d", e.subcommand, e.exitCode)
}
