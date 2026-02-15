//go:build darwin

package runtime

import (
	"context"
	"io"
	"os/exec"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/cruciblehq/crux/kit/crex"
)

// Copies a tar stream into a running container.
//
// On macOS the containerd shim runs inside a Lima VM, so the tar stream is
// piped through limactl shell into a ctr task exec running tar inside the
// container. The stream is extracted at destDir which must already exist.
func containerCopy(_ context.Context, _ *containerd.Client, registry, id string, r io.Reader, destDir string) error {
	l, err := newLima()
	if err != nil {
		return crex.Wrap(ErrContainerCopy, err)
	}

	ctrArgs := []string{
		"-n", registry,
		"-a", containerdGuestSock,
		"task", "exec",
		"--exec-id", nextExecID(),
		id, "tar", "xf", "-", "-C", destDir,
	}

	shellArgs := append([]string{"shell", limaInstanceName, "ctr"}, ctrArgs...)
	cmd := exec.Command(l.limactl, shellArgs...)
	cmd.Stdin = r
	cmd.Env = l.env()

	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return crex.Wrap(ErrContainerCopy, &commandError{
				subcommand: "ctr task exec tar",
				exitCode:   exitErr.ExitCode(),
				output:     string(output),
			})
		}
		return crex.Wrap(ErrContainerCopy, err)
	}

	return nil
}
