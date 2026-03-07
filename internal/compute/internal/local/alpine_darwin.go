//go:build darwin

package local

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
)

const (

	// Machine image.
	imageVersion     = "3.23"                                                                            // Alpine Linux version to use for the VM image.
	imageDownloadURL = "https://dl-cdn.alpinelinux.org/alpine/v%s/releases/%s/alpine-virt-%s.0-%s.qcow2" // Download URL template. Placeholders: version, arch, version, arch.

	// Architecture identifiers for machine image release filenames.
	imageArchARM64 = "aarch64" // ARM64
	imageArchAMD64 = "x86_64"  // AMD64
)

// Architecture identifier for machine image release filenames.
func imageArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return imageArchARM64
	case goarchAMD64:
		return imageArchAMD64
	default:
		return imageArchAMD64
	}
}

// Ensures the machine image is available locally, downloading it on first use.
//
// Returns the absolute path to the cached image.
func ensureImage(ctx context.Context) (string, error) {
	dest := paths.AlpineImage()
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	slog.Info("machine image not found, downloading...",
		"version", imageVersion,
		"arch", imageArch(),
	)

	if err := os.MkdirAll(filepath.Dir(dest), paths.DefaultDirMode); err != nil {
		return "", crex.Wrap(ErrImageDownload, err)
	}

	url := fmt.Sprintf(imageDownloadURL, imageVersion, imageArch(), imageVersion, imageArch())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", crex.Wrap(ErrImageDownload, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", crex.Wrap(ErrImageDownload, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", crex.Wrapf(ErrImageDownload, "unexpected status %d from %s", resp.StatusCode, url)
	}

	slog.Debug("download complete, caching image",
		"path", dest,
	)

	f, err := os.Create(dest)
	if err != nil {
		return "", crex.Wrap(ErrImageDownload, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(dest)
		return "", crex.Wrap(ErrImageDownload, err)
	}

	return dest, nil
}

// Ensures the runtime VM is running, creating it if necessary.
//
// If the VM does not exist, a Lima YAML configuration is generated and
// the VM is created and started in one shot. If the VM exists but is
// stopped, it is resumed. If already running, this is a no-op. Blocks
// until the VM is ready.
func ensureRuntimeRunning(ctx context.Context, cruxdVersion string) error {
	if err := ensureLima(ctx); err != nil {
		return err
	}

	status, err := runtimeStatus(ctx)
	if err != nil {
		return err
	}

	switch status {
	case provider.StateRunning:
		return nil

	case provider.StateStopped:
		if err := limaRun(ctx, "start", "--tty=false", limaInstanceName); err != nil {
			return crex.Wrap(ErrRuntimeStart, err)
		}
		return nil

	case provider.StateNotProvisioned:
		imagePath, err := ensureImage(ctx)
		if err != nil {
			return err
		}
		configPath, err := generateLimaConfig(cruxdVersion, imagePath)
		if err != nil {
			return err
		}
		if err := limaRun(ctx, "start", "--tty=false", "--name="+limaInstanceName, configPath); err != nil {
			return crex.Wrap(ErrRuntimeStart, err)
		}
		return nil

	default:
		return crex.Wrapf(ErrRuntimeStart, "unexpected VM state: %s", status)
	}
}

// Deletes the runtime VM and its disk images.
//
// Blocks until cleanup is complete.
func destroyRuntime(ctx context.Context) error {
	status, err := runtimeStatus(ctx)
	if err != nil {
		return err
	}
	if status == provider.StateNotProvisioned {
		return ErrRuntimeNotCreated
	}

	if err := limaRun(ctx, "delete", "--force", limaInstanceName); err != nil {
		return crex.Wrap(ErrRuntimeDestroy, err)
	}
	return nil
}

// Runs a command inside the runtime VM and captures its output.
func runtimeExec(ctx context.Context, command string, args ...string) (*provider.ExecResult, error) {
	return limaExec(ctx, command, args...)
}

// Queries the current state of the runtime VM.
//
// Maps the Lima instance status to a provider state.
func runtimeStatus(ctx context.Context) (provider.State, error) {
	switch limaInstanceStatus(ctx) {
	case limaStatusRunning:
		return provider.StateRunning, nil
	case limaStatusStopped:
		return provider.StateStopped, nil
	default:
		return provider.StateNotProvisioned, nil
	}
}
