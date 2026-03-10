//go:build darwin

package local

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/archive"
)

const (

	// Lima configuration.
	limaVersion      = "2.0.3" // Lima version to use for the crux VM.
	limaInstanceName = "crux"  // Lima instance name used for the crux VM.

	// Status strings returned by limactl list.
	limaStatusRunning = "Running" // Lima instance is running.
	limaStatusStopped = "Stopped" // Lima instance is stopped.

	// Download URL template for Lima releases. Uses placeholders for version,
	// OS, and architecture.
	limaDownloadURL = "https://github.com/lima-vm/lima/releases/download/v%s/lima-%s-%s-%s.tar.gz"

	// Go GOARCH values.
	goarchARM64 = "arm64" // Apple Silicon
	goarchAMD64 = "amd64" // Intel

	// Architecture identifiers used in Lima YAML configuration.
	limaArchARM64 = "aarch64" // ARM64 (Lima uses aarch64)
	limaArchAMD64 = "x86_64"  // AMD64 (Lima uses x86_64)

	// Architecture identifiers used in Darwin release asset filenames.
	downloadArchARM64 = "arm64"  // Apple Silicon
	downloadArchAMD64 = "x86_64" // Intel
)

// Lima architecture identifier for the YAML config.
func limaArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return limaArchARM64
	case goarchAMD64:
		return limaArchAMD64
	default:
		return limaArchAMD64
	}
}

// Architecture identifier for Darwin release asset URLs.
func limaDownloadArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return downloadArchARM64
	case goarchAMD64:
		return downloadArchAMD64
	default:
		return downloadArchAMD64
	}
}

// Ensures the limactl binary is available, downloading it if necessary.
func ensureLima(ctx context.Context) error {
	if _, err := os.Stat(paths.LimactlBin()); err == nil {
		return nil
	}

	slog.Info("Lima not found, downloading...",
		"version", limaVersion,
		"arch", limaDownloadArch(),
	)

	url := fmt.Sprintf(limaDownloadURL, limaVersion, limaVersion, "Darwin", limaDownloadArch())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.Wrapf(ErrLimaDownload, "unexpected status %d from %s", resp.StatusCode, url)
	}

	slog.Debug("download complete, extracting Lima",
		"version", limaVersion,
	)

	return extractLima(resp.Body, paths.LimaDir())
}

// Extracts the Lima distribution from a gzipped tar archive.
//
// All entries are extracted into the destination directory preserving the
// archive's internal structure and executable permissions. This includes
// the limactl binary and supporting files like guest agents.
func extractLima(r io.Reader, dest string) error {
	if err := archive.ExtractFromReader(r, dest, archive.Gzip); err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}

	if _, err := os.Stat(filepath.Join(dest, "bin", "limactl")); err != nil {
		return crex.Wrapf(ErrLimaDownload, "limactl not found in archive")
	}
	return nil
}

// Runs a command inside the VM and captures its output.
//
// Blocks until the command completes or the context is cancelled. The command
// runs as the default Lima user inside the guest. The caller must ensure the
// VM is running before calling this function.
func limaExec(ctx context.Context, command string, args ...string) (*provider.ExecResult, error) {
	shellArgs := append([]string{"shell", limaInstanceName, command}, args...)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, paths.LimactlBin(), shellArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = limaEnv()

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, crex.Wrap(ErrHostExec, err)
		}
	}

	return provider.NewExecResult(stdout.String(), stderr.String(), exitCode), nil
}

// Runs a limactl subcommand synchronously.
//
// Blocks until the command exits or the context is cancelled. I/O is
// disconnected so limactl does not attempt to interact with the terminal.
// Errors are returned unwrapped; callers are responsible for wrapping with
// the appropriate sentinel.
func limaRun(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, paths.LimactlBin(), args...)
	cmd.Env = limaEnv()
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	return cmd.Run()
}

// Runs a command inside the shared VM via limactl shell.
//
// Unlike [limaExec], this function discards output and is intended for
// fire-and-forget operations like starting or stopping cruxd instances.
func limaShell(ctx context.Context, command string, args ...string) error {
	shellArgs := []string{"shell", "--workdir", "/", limaInstanceName, "--"}
	shellArgs = append(shellArgs, command)
	shellArgs = append(shellArgs, args...)

	cmd := exec.CommandContext(ctx, paths.LimactlBin(), shellArgs...)
	cmd.Env = limaEnv()
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	return cmd.Run()
}

// Queries the Lima instance status string.
//
// Returns the raw status string from limactl (e.g. "Running", "Stopped"),
// or an empty string if the instance does not exist.
func limaInstanceStatus(ctx context.Context) string {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, paths.LimactlBin(), "list", "--format={{.Status}}", limaInstanceName)
	cmd.Stdout = &stdout
	cmd.Env = limaEnv()

	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(stdout.String())
}

// Environment for limactl commands.
//
// LIMA_HOME is set to the crux VM directory so Lima stores its instance
// data alongside other crux state rather than in ~/.lima. PATH and HOME
// are preserved from the current process so that limactl can find system
// tools and resolve user directories.
func limaEnv() []string {
	env := []string{"LIMA_HOME=" + paths.VMDir()}

	appendIfSet := func(key string) {
		if val := os.Getenv(key); val != "" {
			env = append(env, key+"="+val)
		}
	}

	appendIfSet("PATH")
	appendIfSet("HOME")
	appendIfSet("USER")
	appendIfSet("TMPDIR")

	return env
}
