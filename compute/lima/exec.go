package lima

import (
	"context"
	"io"
	"strings"

	"github.com/cruciblehq/utils-go/crex"
)

// Function that runs a command in the VM.
//
// The function receives stdout and stderr writers, the command, and its
// arguments. It returns the process exit code and any error encountered while
// starting the command.
type ExecFunc func(ctx context.Context, stdout, stderr io.Writer, command string, args ...string) (int, error)

// Runs a command in the VM, discarding output. Returns an error on non-zero exit.
func run(ctx context.Context, exec ExecFunc, command string, args ...string) error {
	code, err := exec(ctx, io.Discard, io.Discard, command, args...)
	if err != nil {
		return err
	}
	if code != 0 {
		return crex.Newf(ErrMachineExec, "%q exited with code %d", strings.Join(append([]string{command}, args...), " "), code)
	}
	return nil
}

// Runs a command in the VM and returns its stdout as a string.
func execOutput(ctx context.Context, exec ExecFunc, command string, args ...string) (string, error) {
	var stdout strings.Builder
	code, err := exec(ctx, &stdout, io.Discard, command, args...)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", crex.Newf(ErrMachineExec, "%q exited with code %d", strings.Join(append([]string{command}, args...), " "), code)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Runs a command inside a network namespace in the VM.
func nsRun(ctx context.Context, exec ExecFunc, ns, command string, args ...string) error {
	nsArgs := append([]string{netnsSubcommand, execSubcommand, ns, command}, args...)
	return run(ctx, exec, ipCommand, nsArgs...)
}
