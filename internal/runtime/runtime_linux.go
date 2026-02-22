//go:build linux

package runtime

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/archive"
	"github.com/cruciblehq/crux/internal/paths"
	spec "github.com/cruciblehq/spec/paths"
)

const (

	// How long to wait for the cruxd socket to appear after starting.
	daemonStartTimeout = 30 * time.Second

	// Polling interval when waiting for the cruxd socket.
	daemonPollInterval = 250 * time.Millisecond
)

// Path to the installed cruxd binary.
func cruxdPath() string {
	return filepath.Join(paths.Data(), "bin", "cruxd")
}

// Ensures the cruxd binary is available, downloading it if necessary.
func ensureCruxd() error {
	bin := cruxdPath()
	if _, err := os.Stat(bin); err == nil {
		return nil
	}

	binDir := filepath.Dir(bin)
	if err := os.MkdirAll(binDir, paths.DefaultDirMode); err != nil {
		return crex.Wrap(ErrDaemonInstall, err)
	}

	url := fmt.Sprintf(cruxdDownloadURL, goruntime.GOARCH)
	return downloadCruxd(url, binDir)
}

// Downloads the cruxd release archive and extracts the binary into dest.
func downloadCruxd(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return crex.Wrap(ErrDaemonInstall, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.Wrapf(ErrDaemonInstall, "unexpected status %d from %s", resp.StatusCode, url)
	}

	if err := archive.ExtractFromReader(resp.Body, dest, archive.Gzip); err != nil {
		return crex.Wrap(ErrDaemonInstall, err)
	}

	bin := filepath.Join(dest, "cruxd")
	if err := os.Chmod(bin, paths.DefaultExecMode); err != nil {
		return crex.Wrap(ErrDaemonInstall, err)
	}

	return nil
}

// Checks whether the cruxd socket is reachable.
func isDaemonRunning() bool {
	conn, err := net.Dial("unix", spec.Socket())
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Waits for the cruxd socket to become reachable within the timeout.
func waitForDaemon() error {
	deadline := time.Now().Add(daemonStartTimeout)
	for time.Now().Before(deadline) {
		if isDaemonRunning() {
			return nil
		}
		time.Sleep(daemonPollInterval)
	}
	return crex.UserError("daemon not reachable", "cruxd did not start within the expected time").
		Fallback("Check the cruxd logs for errors.").
		Err()
}

// Starts the container runtime environment.
//
// On Linux this ensures the cruxd binary is installed and starts the daemon
// process. Blocks until the daemon socket is reachable.
func Start() error {
	if isDaemonRunning() {
		return ErrRuntimeAlreadyRunning
	}

	if err := ensureCruxd(); err != nil {
		return err
	}

	if err := os.MkdirAll(spec.Runtime(), paths.DefaultDirMode); err != nil {
		return crex.Wrap(ErrRuntimeStart, err)
	}

	cmd := exec.Command(cruxdPath())
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return crex.Wrap(ErrRuntimeStart, err)
	}

	return waitForDaemon()
}

// Stops the daemon process.
func Stop() error {
	if !isDaemonRunning() {
		return ErrRuntimeNotRunning
	}

	// Signal the daemon via its PID file.
	data, err := os.ReadFile(spec.PIDFile())
	if err != nil {
		return crex.Wrap(ErrRuntimeStop, err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return crex.Wrap(ErrRuntimeStop, err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return crex.Wrap(ErrRuntimeStop, err)
	}

	if err := proc.Signal(os.Interrupt); err != nil {
		return crex.Wrap(ErrRuntimeStop, err)
	}

	return nil
}

// Removes the cruxd binary and runtime state.
func Destroy() error {
	if isDaemonRunning() {
		if err := Stop(); err != nil {
			return err
		}
	}

	os.Remove(cruxdPath())
	os.RemoveAll(spec.Runtime())

	return nil
}

// Queries the current state of the container runtime environment.
func Status() (State, error) {
	if isDaemonRunning() {
		return StateRunning, nil
	}
	if _, err := os.Stat(cruxdPath()); err == nil {
		return StateStopped, nil
	}
	return StateNotCreated, nil
}

// Runs a command directly on the host.
//
// On Linux there is no virtual machine, so the command is executed in the
// current environment.
func Exec(command string, args ...string) (*ExecResult, error) {
	if !isDaemonRunning() {
		return nil, ErrRuntimeNotRunning
	}

	cmd := exec.Command(command, args...)
	stdout, err := cmd.Output()

	exitCode := 0
	stderr := ""
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = string(exitErr.Stderr)
		} else {
			return nil, crex.Wrap(ErrRuntimeExec, err)
		}
	}

	return &ExecResult{
		Stdout:   string(stdout),
		Stderr:   stderr,
		ExitCode: exitCode,
	}, nil
}
