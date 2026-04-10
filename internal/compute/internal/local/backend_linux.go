//go:build linux

package local

import (
	"context"
	"net"
	"os/exec"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Provisions a compute instance.
//
// On Linux containerd runs as a system service. Provisioning verifies
// the socket is reachable.
func provision(_ context.Context, name string, _ resource.Source) error {
	if !isContainerdRunning(name) {
		return ErrHostNotRunning
	}
	return nil
}

// Starts the compute instance.
//
// containerd is managed by the system init, so this only verifies reachability.
func start(_ context.Context, name string) error {
	if !isContainerdRunning(name) {
		return ErrHostNotRunning
	}
	return nil
}

// Stops the compute instance. No-op on Linux — containerd is a system service.
func stop(_ context.Context, _ string) error {
	return nil
}

// Tears down the compute instance. No-op on Linux — containerd is a system service.
func deprovision(_ context.Context, _ string) error {
	return nil
}

// Queries the current state of the compute instance.
func status(_ context.Context, name string) (provider.State, error) {
	if isContainerdRunning(name) {
		return provider.StateRunning, nil
	}
	return provider.StateStopped, nil
}

// Runs a command on the host and captures its output.
func execute(ctx context.Context, _ string, command string, args ...string) (*provider.ExecResult, error) {
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

// Checks whether the containerd socket is reachable.
func isContainerdRunning(name string) bool {
	conn, err := net.Dial("unix", paths.ContainerdSocket(name))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
