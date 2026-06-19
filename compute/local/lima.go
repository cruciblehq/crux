//go:build darwin || linux

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
	"sort"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/cruciblehq/utils-go/archive"
	"github.com/cruciblehq/utils-go/crex"
)

// Default client name used for local runtime paths.
const defaultClientName = "crux"

// Lima configuration.
const (
	limaVersion      = "2.0.3" // Lima version to use for the crux VM.
	limaInstanceName = "crux"  // Lima instance name used for the crux VM.
)

// Path to the vendored Lima installation directory.
func limaDir() string {
	return filepath.Join(xdg.DataHome, defaultClientName, "lima")
}

// Path to the vendored limactl binary.
func limactlBin() string {
	return filepath.Join(limaDir(), "bin", "limactl")
}

// Path to the VM data directory.
func vmDir() string {
	return filepath.Join(xdg.DataHome, defaultClientName, "vm")
}

// Download URL template for Lima releases. Uses placeholders for version,
// OS, and architecture.
const limaDownloadURL = "https://github.com/lima-vm/lima/releases/download/v%s/lima-%s-%s-%s.tar.gz"

// OS names used in Lima release asset URLs.
const (
	limaOSDarwin = "Darwin" // macOS release asset name.
	limaOSLinux  = "Linux"  // Linux release asset name.
)

// Go GOARCH values.
const (
	goarchARM64 = "arm64" // 64-bit ARM.
	goarchAMD64 = "amd64" // 64-bit x86.
)

// Architecture identifiers used in Lima YAML configuration.
const (
	limaArchARM64 = "aarch64" // ARM64 (Lima uses aarch64).
	limaArchAMD64 = "x86_64"  // AMD64 (Lima uses x86_64).
)

// Maximum time to wait for child process I/O pipes to drain after context
// cancellation kills the process.
const commandWaitDelay = 5 * time.Second

// Status strings returned by limactl list.
const (
	limaStatusRunning = "Running" // Lima instance is running.
	limaStatusStopped = "Stopped" // Lima instance is stopped.
)

// Shared recovery guidance for local Lima command failures.
const recoveryRestartLocalEnvironment = "Restart the local environment and try again."

// Process exit codes returned by limaExec.
const (
	exitCodeSuccess = 0  // Command exited cleanly.
	exitCodeError   = -1 // Command could not be started or its exit code is unavailable.
)

// OS name used in Lima release asset URLs (e.g. "Darwin", "Linux").
func limaOS() string {
	switch runtime.GOOS {
	case "darwin":
		return limaOSDarwin
	case "linux":
		return limaOSLinux
	default:
		return runtime.GOOS
	}
}

// Lima architecture identifier for the VM YAML config.
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

// Architecture identifier for release asset URLs.
//
// Darwin and Linux use different names for the ARM64 architecture in release
// tarballs: Darwin uses "arm64" while Linux uses "aarch64". AMD64 is "x86_64"
// on both.
func limaDownloadArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		if runtime.GOOS == "linux" {
			return limaArchARM64 // "aarch64" on Linux
		}
		return goarchARM64 // "arm64" on Darwin
	default:
		return limaArchAMD64 // "x86_64"
	}
}

// Builds the download URL for the appropriate Lima release based on the host
// OS and architecture.
//
// Lima release asset names include the OS and architecture, which are derived
// from the host runtime. For example, on an AMD64 Linux host, this returns the
// URL for the "lima-2.0.3-Linux-x86_64.tar.gz" release asset.
func limaURL() string {
	return fmt.Sprintf(limaDownloadURL, limaVersion, limaVersion, limaOS(), limaDownloadArch())
}

// Ensures the limactl binary is available.
//
// Checks whether limactl is already present in the expected location. If not,
// it downloads the release tarball for the host OS and architecture, extracts
// it, and verifies the limactl binary is present.
func ensureLima(ctx context.Context) error {
	const description = "cannot download Lima"

	if _, err := os.Stat(limactlBin()); err == nil {
		return nil
	}

	slog.Info("Lima not found, downloading...",
		"version", limaVersion,
		"arch", limaDownloadArch(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, limaURL(), nil)
	if err != nil {
		return crex.SystemError(description, "failed to build the download request").
			Recovery("If the problem persists, report it to the Crucible team.").
			Cause(crex.Wrap(ErrLimaDownload, err)).
			Err()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return crex.SystemError(description, "the download request failed").
			Recovery("Check your network connection and try again.").
			Cause(crex.Wrap(ErrLimaDownload, err)).
			Err()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.SystemErrorf(description, "the download server returned status %d", resp.StatusCode).
			Recovery("Check your network connection and try again.").
			Cause(crex.Newf(ErrLimaDownload, "unexpected status %d from %s", resp.StatusCode, limaURL())).
			Err()
	}

	slog.Debug("download complete, extracting Lima",
		"version", limaVersion,
	)

	return extractLima(resp.Body, limaDir())
}

// Extracts the Lima distribution from a gzipped tar archive.
//
// All entries are extracted into the destination directory, preserving the
// archive's internal structure and executable permissions. This includes the
// limactl binary and supporting files like guest agents.
func extractLima(r io.Reader, dest string) error {
	if err := archive.ExtractFromReader(r, dest, archive.Gzip); err != nil {
		return crex.SystemError("cannot install Lima", "failed to extract the Lima archive").
			Recovery("Free up disk space, then try again.").
			Cause(crex.Wrap(ErrLimaDownload, err)).
			Err()
	}

	if _, err := os.Stat(limactlBin()); err != nil {
		return crex.SystemError("cannot install Lima", "the Lima archive did not contain limactl").
			Recovery("If the problem persists, report it to the Crucible team.").
			Cause(crex.Newf(ErrLimaDownload, "limactl not found in archive")).
			Err()
	}
	return nil
}

// Runs a command inside the VM.
//
// Blocks until the command completes or the context is cancelled. The command
// runs as the default Lima user inside the guest. The caller must ensure the
// VM is running before calling this function.
func limaGuestExec(ctx context.Context, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	shellArgs := append([]string{"shell", limaInstanceName, command}, args...)

	cmd := exec.CommandContext(ctx, limactlBin(), shellArgs...)
	cmd.WaitDelay = commandWaitDelay
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = limaEnv()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return exitCodeError, crex.SystemError("cannot run command in local environment", "the command could not be started in the virtual machine").
			Recovery(recoveryRestartLocalEnvironment).
			Cause(crex.Wrap(ErrHostExec, err)).
			Err()
	}

	return exitCodeSuccess, nil
}

// Lists all Lima instances managed by crux.
//
// Runs limactl list to enumerate all instances in the crux LIMA_HOME directory.
// Names are returned sorted lexicographically. Returns an empty slice when no
// instances have been provisioned.
func limaList(ctx context.Context) ([]string, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, limactlBin(), "list", "--format={{.Name}}")
	cmd.WaitDelay = commandWaitDelay
	cmd.Stdout = &stdout
	cmd.Env = limaEnv()
	if err := cmd.Run(); err != nil {
		return nil, crex.SystemError("cannot list local environments", "limactl failed to list instances").
			Recovery(recoveryRestartLocalEnvironment).
			Cause(crex.Wrap(ErrLimaCtl, err)).
			Err()
	}
	names := strings.Fields(stdout.String())
	sort.Strings(names)
	return names, nil
}

// Extracts a tar archive into the root filesystem of the named Lima instance.
//
// Pipes r into "tar -x -C /" running inside the instance via limactl shell,
// which uses the existing SSH connection. Entries in the archive are applied
// as absolute paths, preserving permissions, ownership, and timestamps.
func limaCopyArchive(ctx context.Context, name string, r io.Reader) error {
	cmd := exec.CommandContext(ctx, limactlBin(), "shell", name, "tar", "-x", "-C", "/")
	cmd.WaitDelay = commandWaitDelay
	cmd.Stdin = r
	cmd.Env = limaEnv()
	if err := cmd.Run(); err != nil {
		return crex.SystemError("cannot copy files to local environment", "extracting the archive in the virtual machine failed").
			Recovery(recoveryRestartLocalEnvironment).
			Cause(crex.Wrap(ErrHostExec, err)).
			Err()
	}
	return nil
}

// Runs a limactl subcommand synchronously.
//
// Blocks until the command exits or the context is cancelled. Lima stderr is
// captured and appended to any non-zero exit error so callers can see what
// Lima reported without the output reaching the user's terminal.
func limactlRun(ctx context.Context, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, limactlBin(), args...)
	cmd.WaitDelay = commandWaitDelay
	cmd.Stderr = &stderr
	cmd.Env = limaEnv()

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return crex.Wrapf(ErrLimaCtl, err, "%s", msg)
		}
		return crex.Wrap(ErrLimaCtl, err)
	}
	return nil
}

// Runs a limactl subcommand synchronously with TTY disabled.
//
// Some limactl commands (e.g. start) emit interactive output (like progress
// bars) that does not render well when captured. This is a wrapper around
// limactlRun that adds the --tty=false flag to disable interactive output
// for commands where this is an issue.
func limactlRunNoTTY(ctx context.Context, args ...string) error {
	return limactlRun(ctx, append([]string{"--tty=false"}, args...)...)
}

// Queries the Lima instance status string.
//
// Returns the raw status string from limactl (e.g. "Running", "Stopped"), or
// an empty string if the instance does not exist.
func limaInstanceStatus(ctx context.Context) string {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, limactlBin(), "list", "--format={{.Status}}", limaInstanceName)
	cmd.WaitDelay = commandWaitDelay
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
// tools and resolve user directories. USER and TMPDIR are also preserved.
func limaEnv() []string {
	env := []string{fmt.Sprintf("LIMA_HOME=%s", vmDir())}

	appendIfSet := func(key string) {
		if val := os.Getenv(key); val != "" {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	appendIfSet("PATH")
	appendIfSet("HOME")
	appendIfSet("USER")
	appendIfSet("TMPDIR")

	return env
}
