//go:build linux

package local

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/archive"
	"github.com/cruciblehq/spec/protocol"
)

const (

	// Cruxd version to provision on Linux.
	//
	// The cruxd binary is downloaded directly from the GitHub release. Bump
	// this when adopting a new cruxd release.
	cruxdVersion = "0.3.2"

	// Download URL template for cruxd releases.
	cruxdDownloadURL = "https://github.com/cruciblehq/cruxd/releases/download/v%s/cruxd-linux-%s.tar.gz"
)

// Ensures the cruxd binary is available, downloading it if necessary.
func ensureCruxd(ctx context.Context) error {
	bin := paths.CruxdBin()
	if _, err := os.Stat(bin); err == nil {
		return nil
	}

	binDir := filepath.Dir(bin)
	if err := os.MkdirAll(binDir, paths.DefaultDirMode); err != nil {
		return crex.Wrap(ErrCruxdInstall, err)
	}

	return downloadCruxd(ctx, fmt.Sprintf(cruxdDownloadURL, cruxdVersion, runtime.GOARCH), binDir)
}

// Downloads the cruxd release archive and extracts the binary into dest.
func downloadCruxd(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return crex.Wrap(ErrCruxdInstall, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return crex.Wrap(ErrCruxdInstall, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.Wrapf(ErrCruxdInstall, "unexpected status %d from %s", resp.StatusCode, url)
	}

	if err := archive.ExtractFromReader(resp.Body, dest, archive.Gzip); err != nil {
		return crex.Wrap(ErrCruxdInstall, err)
	}

	bin := filepath.Join(dest, "cruxd")
	if err := os.Chmod(bin, paths.DefaultExecMode); err != nil {
		return crex.Wrap(ErrCruxdInstall, err)
	}

	return nil
}

// Checks whether the cruxd socket for an instance is reachable.
func isCruxdRunning(name string) bool {
	conn, err := net.Dial("unix", paths.CruxdSocket(name))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Starts the cruxd process for an instance.
//
// The caller must ensure the cruxd binary is available before calling this
// function (see [ensureCruxd]). Readiness is detected via ready-fd: a pipe
// is created, the write end is passed to cruxd as an extra file descriptor
// via --ready-fd, and this function blocks reading the read end. cruxd writes
// a CmdOK message once the socket is bound, unblocking the reader. If cruxd
// exits before signaling, the read returns EOF and an error is raised.
func startCruxd(name string) error {
	if isCruxdRunning(name) {
		return ErrHostAlreadyRunning
	}

	if err := os.MkdirAll(paths.CruxdInstanceDir(name), paths.DefaultDirMode); err != nil {
		return crex.Wrap(ErrHostStart, err)
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return crex.Wrap(ErrHostStart, err)
	}
	defer pr.Close()

	// ExtraFiles[0] becomes fd 3 in the child process.
	cmd := exec.Command(
		paths.CruxdBin(), "start",
		"--socket", paths.CruxdSocket(name),
		"--pid-file", paths.CruxdPIDFile(name),
		"--ready-fd", "3",
	)
	cmd.ExtraFiles = []*os.File{pw}
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		pw.Close()
		return crex.Wrap(ErrHostStart, err)
	}
	// Close the write end in the parent so reads get EOF if cruxd exits.
	pw.Close()

	line := make([]byte, 128)
	n, readErr := pr.Read(line)
	env, _, decErr := protocol.Decode(line[:n])
	if readErr != nil || decErr != nil || env.Command != protocol.CmdOK {
		return crex.Wrapf(ErrHostStart, "cruxd did not signal readiness")
	}

	return nil
}

// Signals the cruxd process to stop and returns its PID.
func stopCruxd(name string) (int, error) {
	if !isCruxdRunning(name) {
		return 0, ErrHostNotRunning
	}

	data, err := os.ReadFile(paths.CruxdPIDFile(name))
	if err != nil {
		return 0, crex.Wrap(ErrHostStop, err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, crex.Wrap(ErrHostStop, err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, crex.Wrap(ErrHostStop, err)
	}

	if err := proc.Signal(os.Interrupt); err != nil {
		return 0, crex.Wrap(ErrHostStop, err)
	}

	return pid, nil
}
