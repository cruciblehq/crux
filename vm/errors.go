package vm

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupportedPlatform = errors.New("vm is only supported on macOS")
	ErrLimaNotFound        = errors.New("lima binary not found")
	ErrLimaDownload        = errors.New("failed to download lima")
	ErrVMNotCreated        = errors.New("vm has not been created")
	ErrVMAlreadyRunning    = errors.New("vm is already running")
	ErrVMNotRunning        = errors.New("vm is not running")
	ErrVMStart             = errors.New("failed to start vm")
	ErrVMStop              = errors.New("failed to stop vm")
	ErrVMExec              = errors.New("failed to execute command in vm")
	ErrVMConfig            = errors.New("failed to generate vm configuration")
)

// Error from a failed limactl command.
//
// Holds the subcommand name, exit code, and raw output separately so callers
// can present a clean message to users while preserving diagnostic detail for
// debug logging.
type CommandError struct {
	Subcommand string // The limactl subcommand that failed (e.g. "start").
	ExitCode   int    // Process exit code.
	Output     string // Combined stdout/stderr from the process.
}

// Returns a user-facing summary without leaking raw CLI output.
func (e *CommandError) Error() string {
	return fmt.Sprintf("limactl %s exited with code %d", e.Subcommand, e.ExitCode)
}
