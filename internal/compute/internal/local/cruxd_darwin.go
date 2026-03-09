//go:build darwin

package local

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/protocol"
)

// Starts a cruxd process inside the VM for the given instance.
//
// The cruxd process listens on the instance's socket path, which resides on
// the virtiofs mount shared between host and guest. This allows the host to
// connect directly without SSH port forwarding. Readiness is detected via
// ready-fd: --ready-fd 1 tells cruxd to write a [protocol.CmdOK] envelope to
// stdout once the socket is bound. limactl forwards the guest's stdout to the
// host through the SSH channel, so this function reads the signal directly
// from the process pipe. The limactl process remains running as the session
// holder for cruxd and exits when cruxd is stopped.
func startCruxd(ctx context.Context, name string) error {
	socketPath := paths.CruxdSocket(name)
	pidPath := paths.CruxdPIDFile(name)

	// Remove stale artifacts from a previous run (crash, forced stop).
	os.Remove(socketPath)
	os.Remove(pidPath)

	cmd := exec.CommandContext(ctx, paths.LimactlBin(),
		"shell", limaInstanceName, "--",
		"sudo", "/usr/local/bin/cruxd", "start",
		"--socket", socketPath,
		"--pid-file", pidPath,
		"--ready-fd", "1",
	)
	cmd.Env = limaEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return crex.Wrap(ErrRuntimeStart, err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return crex.Wrap(ErrRuntimeStart, err)
	}

	// Block until cruxd writes a ready envelope to stdout or the process exits.
	line, readErr := bufio.NewReader(stdout).ReadBytes('\n')

	env, _, decErr := protocol.Decode(line)
	if readErr != nil || decErr != nil || env.Command != protocol.CmdOK {
		cmd.Process.Kill()
		cmd.Wait()
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return crex.Wrapf(ErrRuntimeStart, "%s", detail)
		}
		return crex.Wrapf(ErrRuntimeStart, "cruxd did not signal readiness")
	}

	// cruxd is ready. The limactl process remains running as the session
	// holder — it exits automatically when cruxd is stopped.
	go cmd.Wait()
	return nil
}

// Stops the cruxd process for the given instance.
//
// The guest PID is read from the PID file on the virtiofs mount and
// signalled via limactl shell. The socket and PID file are removed
// after the process is stopped.
func stopCruxd(ctx context.Context, name string) error {
	pidPath := paths.CruxdPIDFile(name)

	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrRuntimeNotRunning
		}
		return crex.Wrap(ErrRuntimeStop, err)
	}

	pid := strings.TrimSpace(string(data))
	result, err := limaExec(ctx, "sudo", "kill", pid)
	if err != nil {
		return crex.Wrap(ErrRuntimeStop, err)
	}
	if result.ExitCode != 0 {
		return crex.Wrapf(ErrRuntimeStop, "%s", strings.TrimSpace(result.Stderr))
	}

	os.Remove(paths.CruxdSocket(name))
	os.Remove(pidPath)
	return nil
}
