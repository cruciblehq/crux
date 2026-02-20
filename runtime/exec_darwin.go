//go:build darwin

package runtime

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/cruciblehq/crex"
)

const (

	// Path to the containerd socket inside the guest VM.
	containerdGuestSock = "/run/containerd/containerd.sock"
)

// Runs a command inside a container and captures its output.
//
// On macOS the containerd shim runs inside a Lima VM. The containerd gRPC API
// (Task.Exec) uses FIFOs for process IO, but FIFOs do not work across the VM
// boundary. Instead of using the containerd API, this routes through limactl
// to invoke `ctr task exec` inside the guest where the shim, FIFOs, and client
// all share the same kernel.
func containerExec(_ context.Context, _ *containerd.Client, registry, id string, opts ExecOptions, command string, args ...string) (*ExecResult, error) {
	l, err := newLima()
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	ctrArgs := []string{
		"-n", registry,
		"-a", containerdGuestSock,
		"task", "exec",
		"--exec-id", nextExecID(),
	}

	if opts.Workdir != "" {
		ctrArgs = append(ctrArgs, "--cwd", opts.Workdir)
	}
	for _, env := range opts.Env {
		ctrArgs = append(ctrArgs, "--env", env)
	}

	ctrArgs = append(ctrArgs, id, command)
	ctrArgs = append(ctrArgs, args...)

	result, err := l.exec("ctr", ctrArgs...)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	return result, nil
}
