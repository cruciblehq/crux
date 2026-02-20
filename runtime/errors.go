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
