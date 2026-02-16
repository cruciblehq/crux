package runtime

import (
	"errors"
	"fmt"
)

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

	// Lima errors (macOS only).
	ErrLimaDownload = errors.New("failed to download lima")

	// Containerd errors.
	ErrContainerd         = errors.New("failed to connect to containerd")
	ErrContainerdDownload = errors.New("failed to download containerd")
	ErrContainerdNotFound = errors.New("containerd binary not found in archive")
	ErrContainerdConfig   = errors.New("failed to generate containerd configuration")
	ErrContainerdStart    = errors.New("failed to start containerd")
	ErrContainerdStop     = errors.New("failed to stop containerd")
	ErrContainerdTimeout  = errors.New("timed out waiting for containerd to start")
	ErrContainerdExited   = errors.New("containerd exited unexpectedly")

	// Image errors.
	ErrImageFileOpen      = errors.New("failed to open image file")
	ErrImageImport        = errors.New("failed to import image into containerd")
	ErrImageEmpty         = errors.New("archive contains no images")
	ErrImageMultiple      = errors.New("archive contains multiple images")
	ErrImageDestroy       = errors.New("failed to destroy image")
	ErrImageExport        = errors.New("failed to export image")

	// Container errors.
	ErrContainerStart   = errors.New("failed to start container")
	ErrContainerStop    = errors.New("failed to stop container")
	ErrContainerDestroy = errors.New("failed to destroy container")
	ErrContainerExec    = errors.New("failed to execute command in container")
	ErrContainerCopy    = errors.New("failed to copy into container")
	ErrContainerCopyOut = errors.New("failed to copy from container")
	ErrContainerStatus  = errors.New("failed to query container status")
	ErrContainerCommit  = errors.New("failed to commit container")
)

// A failed runtime command.
//
// The subcommand name, exit code, and raw output are stored separately so
// callers can present a clean message to users while preserving diagnostic
// detail for debug logging.
type commandError struct {
	subcommand string // The subcommand that failed (e.g. "start").
	exitCode   int    // Process exit code.
	output     string // Combined stdout/stderr from the process.
}

// User-facing summary without raw CLI output.
func (e *commandError) Error() string {
	return fmt.Sprintf("runtime command %q exited with code %d", e.subcommand, e.exitCode)
}
