//go:build linux

package local

import (
	"context"
	"os"
	"os/exec"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Provisions a cruxd instance.
func provision(ctx context.Context, name string, _ resource.Source) error {
	if err := ensureCruxd(ctx); err != nil {
		return err
	}
	if err := startCruxd(name); err != nil {
		return err
	}
	return nil
}

// Starts the cruxd process.
func start(_ context.Context, name string) error {
	if err := startCruxd(name); err != nil {
		return err
	}
	return nil
}

// Signals the cruxd process to stop and waits for exit.
func stop(_ context.Context, name string) error {
	pid, err := stopCruxd(name)
	if err != nil {
		return err
	}
	waitForProcessExit(pid)
	return nil
}

// Tears down the cruxd process and removes all instance state.
func deprovision(_ context.Context, name string) error {
	if isCruxdRunning(name) {
		pid, err := stopCruxd(name)
		if err != nil {
			return err
		}
		waitForProcessExit(pid)
	}

	os.RemoveAll(paths.CruxdInstanceDir(name))
	return nil
}

// Queries the current state of the cruxd process.
func status(_ context.Context, name string) (provider.State, error) {
	if isCruxdRunning(name) {
		return provider.StateRunning, nil
	}
	if _, err := os.Stat(paths.CruxdInstanceDir(name)); err == nil {
		return provider.StateStopped, nil
	}
	return provider.StateNotProvisioned, nil
}

// Runs a command on the host and captures its output.
func execute(ctx context.Context, name string, command string, args ...string) (*provider.ExecResult, error) {
	if !isCruxdRunning(name) {
		return nil, ErrHostNotRunning
	}

	cmd := exec.CommandContext(ctx, command, args...)
	stdout, err := cmd.Output()

	exitCode := 0
	stderr := ""
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = string(exitErr.Stderr)
		} else {
			return nil, crex.Wrap(ErrHostExec, err)
		}
	}

	return provider.NewExecResult(string(stdout), stderr, exitCode), nil
}
