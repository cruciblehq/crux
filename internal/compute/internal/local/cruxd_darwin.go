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
	specpaths "github.com/cruciblehq/spec/paths"
	"github.com/cruciblehq/spec/protocol"
)

const (

	// File mode applied to the cruxd guest socket after startup.
	//
	// cruxd restricts its socket to owner+group, but Lima's port-forwarding
	// user is neither owner nor in the group. The socket must be
	// world-accessible so Lima can tunnel connections from the host.
	cruxdSocketMode = "0666"
)

// Starts a cruxd process inside the VM for the given instance.
//
// cruxd binds its socket on a guest-local path (/run/cruxd/...) because
// Unix domain sockets on virtiofs are visible on both sides but the kernel
// socket state does not cross the boundary. Lima's portForwards section
// tunnels the guest socket to the host over SSH, so the host can dial its
// local path transparently. The PID file is written to the virtiofs mount
// so the host can read it directly for stop/status operations.
//
// Readiness is detected via ready-fd: --ready-fd 1 tells cruxd to write a
// [protocol.CmdOK] envelope to stdout once the socket is bound. limactl
// forwards the guest's stdout to the host through the SSH channel, so this
// function reads the signal directly from the process pipe. The limactl
// process remains running as the session holder for cruxd and exits when
// cruxd is stopped.
func startCruxd(ctx context.Context, name string) error {
	guestSocket := specpaths.Socket(name)
	pidPath := paths.CruxdPIDFile(name)

	// Remove stale PID file from a previous run (crash, forced stop). The host
	// socket is managed by Lima's port forwarding and cleaned up automatically
	// when the guest socket disappears.
	os.Remove(pidPath)

	cmd := exec.CommandContext(ctx, paths.LimactlBin(),
		"shell", limaInstanceName, "--",
		"sudo", "/usr/local/bin/cruxd", "start",
		"--socket", guestSocket,
		"--pid-file", pidPath,
		"--ready-fd", "1",
	)
	cmd.WaitDelay = commandWaitDelay
	cmd.Env = limaEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return crex.Wrap(ErrHostStart, err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return crex.Wrap(ErrHostStart, err)
	}

	// Block until cruxd writes a ready envelope to stdout or the process exits.
	line, readErr := bufio.NewReader(stdout).ReadBytes('\n')

	env, _, decErr := protocol.Decode(line)
	if readErr != nil || decErr != nil || env.Command != protocol.CmdOK {
		cmd.Process.Kill()
		cmd.Wait()
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return crex.Wrapf(ErrHostStart, "%s", detail)
		}
		return crex.Wrapf(ErrHostStart, "cruxd did not signal readiness")
	}

	// cruxd is ready. Open the socket to the Lima SSH user so Lima's port
	// forwarding can tunnel connections from the host. cruxd restricts the
	// socket to owner+group by default, but the Lima user is neither.
	if _, err := limaExec(ctx, "sudo", "chmod", cruxdSocketMode, guestSocket); err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return crex.Wrap(ErrHostStart, err)
	}

	// The limactl process remains running as the session
	// holder — it exits automatically when cruxd is stopped.
	go cmd.Wait()
	return nil
}

// Stops the cruxd process for the given instance.
//
// The guest PID is read from the PID file on the virtiofs mount and signalled
// via limactl. The socket and PID file are removed after the process is stopped.
func stopCruxd(ctx context.Context, name string) error {
	pidPath := paths.CruxdPIDFile(name)

	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrHostNotRunning
		}
		return crex.Wrap(ErrHostStop, err)
	}

	pid := strings.TrimSpace(string(data))
	result, err := limaExec(ctx, "sudo", "kill", pid)
	if err != nil {
		return crex.Wrap(ErrHostStop, err)
	}
	if result.ExitCode != 0 {
		return crex.Wrapf(ErrHostStop, "%s", strings.TrimSpace(result.Stderr))
	}

	// The host socket is managed by Lima's port forwarding and is cleaned
	// up automatically when the guest socket disappears.
	os.Remove(pidPath)
	return nil
}
