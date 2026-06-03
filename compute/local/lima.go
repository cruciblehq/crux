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
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/cruciblehq/crux/archive"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
)

const (

	// Lima configuration.
	limaVersion      = "2.0.3" // Lima version to use for the crux VM.
	limaInstanceName = "crux"  // Lima instance name used for the crux VM.

	// Download URL template for Lima releases. Uses placeholders for version,
	// OS, and architecture.
	limaDownloadURL = "https://github.com/lima-vm/lima/releases/download/v%s/lima-%s-%s-%s.tar.gz"

	// OS names used in Lima release asset URLs.
	limaOSDarwin = "Darwin" // macOS release asset name.
	limaOSLinux  = "Linux"  // Linux release asset name.

	// Go GOARCH values.
	goarchARM64 = "arm64" // 64-bit ARM.
	goarchAMD64 = "amd64" // 64-bit x86.

	// Architecture identifiers used in Lima YAML configuration.
	limaArchARM64 = "aarch64" // ARM64 (Lima uses aarch64).
	limaArchAMD64 = "x86_64"  // AMD64 (Lima uses x86_64).

	// Maximum time to wait for child process I/O pipes to drain after context
	// cancellation kills the process.
	commandWaitDelay = 5 * time.Second

	// Status strings returned by limactl list.
	limaStatusRunning = "Running" // Lima instance is running.
	limaStatusStopped = "Stopped" // Lima instance is stopped.

	// Process exit codes returned by limaExec.
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
	if _, err := os.Stat(files.LimactlBin()); err == nil {
		return nil
	}

	slog.Info("Lima not found, downloading...",
		"version", limaVersion,
		"arch", limaDownloadArch(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, limaURL(), nil)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crex.Wrapf(ErrLimaDownload, "unexpected status %d from %s", resp.StatusCode, limaURL())
	}

	slog.Debug("download complete, extracting Lima",
		"version", limaVersion,
	)

	return extractLima(resp.Body, files.LimaDir())
}

// Extracts the Lima distribution from a gzipped tar archive.
//
// All entries are extracted into the destination directory, preserving the
// archive's internal structure and executable permissions. This includes the
// limactl binary and supporting files like guest agents.
func extractLima(r io.Reader, dest string) error {
	if err := archive.ExtractFromReader(r, dest, archive.Gzip); err != nil {
		return crex.Wrap(ErrLimaDownload, err)
	}

	if _, err := os.Stat(files.LimactlBin()); err != nil {
		return crex.Wrapf(ErrLimaDownload, "limactl not found in archive")
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

	cmd := exec.CommandContext(ctx, files.LimactlBin(), shellArgs...)
	cmd.WaitDelay = commandWaitDelay
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = limaEnv()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return exitCodeError, crex.Wrap(ErrHostExec, err)
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
	cmd := exec.CommandContext(ctx, files.LimactlBin(), "list", "--format={{.Name}}")
	cmd.WaitDelay = commandWaitDelay
	cmd.Stdout = &stdout
	cmd.Env = limaEnv()
	if err := cmd.Run(); err != nil {
		return nil, crex.Wrap(ErrLimaCtl, err)
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
	cmd := exec.CommandContext(ctx, files.LimactlBin(), "shell", name, "tar", "-x", "-C", "/")
	cmd.WaitDelay = commandWaitDelay
	cmd.Stdin = r
	cmd.Env = limaEnv()
	if err := cmd.Run(); err != nil {
		return crex.Wrap(ErrHostExec, err)
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
	cmd := exec.CommandContext(ctx, files.LimactlBin(), args...)
	cmd.WaitDelay = commandWaitDelay
	cmd.Stderr = &stderr
	cmd.Env = limaEnv()

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return crex.Wrapf(ErrLimaCtl, "%s: %s", err, msg)
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
	cmd := exec.CommandContext(ctx, files.LimactlBin(), "list", "--format={{.Status}}", limaInstanceName)
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
	env := []string{"LIMA_HOME=" + files.VMDir()}

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
