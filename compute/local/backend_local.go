//go:build darwin || linux

package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/hub"
	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/utils-go/crex"
)

// Default machine image info.
const (
	machineType      = "machine"         // Registry resource type for the machine image.
	machineNamespace = "crucible"        // Registry namespace for the machine image.
	machineName      = "machine-default" // Registry resource name for the machine image.
	machineVersion   = "0.2.0"           // Pinned machine image version.
	machineExtension = ".qcow2"          // Disk image file extension.
)

// Host interface Lima forwards cruxd's port to.
const cruxdHost = "127.0.0.1"

// Ensures the default machine disk image is available in the local cache.
//
// Downloads from a given registry if not already cached. Returns the local
// filesystem path to the image file.
func ensureMachineImage(ctx context.Context, registryURL *url.URL) (string, error) {
	path, err := cachedMachineImagePath()
	if errors.Is(err, ErrMachineImageMissing) {
		slog.Info("machine image not in cache, downloading...",
			"namespace", machineNamespace,
			"name", machineName,
			"version", machineVersion,
		)
		if err := fetchMachineImage(ctx, registryURL); err != nil {
			return "", err
		}
		path, err = cachedMachineImagePath()
	}
	if err != nil {
		return "", err
	}
	return path, nil
}

// Local implementation of [provider.Backend.UploadImage].
//
// The local provider uses the image directly from the local filesystem, so this
// function only verifies the file exists and is accessible. The returned path
// is used as the image ID passed to Lima during provisioning. Other providers
// copy the image file to remote storage and return an image ID.
func uploadImage(_ context.Context, path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", crex.SystemError("cannot access machine image", "the machine image file is missing or unreadable").
			Recovery("Download the machine image again and retry.").
			Cause(crex.Wrap(ErrImageUpload, err)).
			Err()
	}
	return path, nil
}

// Provisions a compute instance from the given disk image.
//
// imageID is the local filesystem path to a QCOW2 disk image, as returned by
// [uploadImage]. enrollment is written into the VM so cruxd can enroll.
// Returns [ErrHostAlreadyProvisioned] if an instance already exists. The VM is
// created and started.
func provision(ctx context.Context, _, imageID string, enrollment []byte) error {
	if err := ensureLima(ctx); err != nil {
		return err
	}
	if hostStatus(ctx) != StateNotProvisioned {
		return ErrHostAlreadyProvisioned
	}
	return createAndStartHost(ctx, imageID, enrollment)
}

// Starts the VM.
//
// The VM must already be provisioned. If already running this is a no-op, and
// if it exists but is stopped, it is resumed. Returns [ErrHostNotCreated] if
// the VM has not been provisioned.
func start(ctx context.Context, _ string) error {
	state := hostStatus(ctx)
	if state == StateRunning {
		return nil
	}
	if state == StateNotProvisioned {
		return ErrHostNotCreated
	}

	if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
		return crex.SystemError("cannot start local environment", "the local virtual machine failed to start").
			Recovery("Recreate the local virtual machine and retry.").
			Cause(crex.Wrap(ErrHostStart, err)).
			Err()
	}
	return nil
}

// Stops the VM.
//
// Returns [ErrHostNotRunning] if the VM is not currently running. limactl stop
// sends a graceful shutdown signal to the guest, giving containerd a chance to
// clean up before the VM halts.
func stop(ctx context.Context, _ string) error {
	state := hostStatus(ctx)
	if state != StateRunning {
		return ErrHostNotRunning
	}

	if err := limactlRun(ctx, "stop", limaInstanceName); err != nil {
		return crex.SystemError("cannot stop local environment", "the local virtual machine failed to stop").
			Recovery("Retry after recreating the local environment.").
			Cause(crex.Wrap(ErrHostStop, err)).
			Err()
	}
	return nil
}

// Tears down the instance and destroys the VM.
//
// After this, the VM no longer appears in limactl list and all resources it
// consumed are freed. If the VM exists, it is terminated and deleted along
// with its disk image. Returns nil if the VM does not exist.
func deprovision(ctx context.Context, _ string) error {
	err := destroyHost(ctx)
	if errors.Is(err, ErrHostNotCreated) {
		return nil
	}
	return err
}

// Queries the current state of the VM.
//
// Reports only whether the VM runs. Whether cruxd is reachable is checked by
// the caller over mutual TLS.
func status(ctx context.Context, _ string) (State, error) {
	return hostStatus(ctx), nil
}

// Runs a command inside the host VM.
//
// Used for setup and teardown tasks that must run on the host. The output is
// streamed to stdout and stderr as it is produced. Returns the exit code and
// a nil error when the process exits normally, regardless of exit code. Also
// returns a non-nil error only if the command could not be started or the
// context was cancelled before the process completed.
func execute(ctx context.Context, _ string, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return hostExec(ctx, stdout, stderr, command, args...)
}

// Lists all instances managed by the local provider.
func list(ctx context.Context) ([]string, error) {
	return limaList(ctx)
}

// Sends a tar archive to the named instance and applies it to the host filesystem.
func copyArchive(ctx context.Context, name string, r io.Reader) error {
	return limaCopyArchive(ctx, name, r)
}

// Creates the Lima instance from the given disk image and starts it.
//
// Lima 2.x create does not auto-start; an explicit start call is required.
// imagePath is the local filesystem path to a QCOW2 disk image. enrollment is
// written into the VM config. Returns once the VM runs, before cruxd is
// enrolled.
func createAndStartHost(ctx context.Context, imagePath string, enrollment []byte) error {
	const description = "cannot create local environment"

	configPath, err := generateLimaConfig(imagePath, enrollment)
	if err != nil {
		return crex.SystemError(description, "failed to generate the virtual machine configuration").
			Recovery("Regenerate the local virtual machine configuration and retry.").
			Cause(crex.Wrap(ErrHostCreate, err)).
			Err()
	}

	if err := limactlRunNoTTY(ctx, "create", fmt.Sprintf("--name=%s", limaInstanceName), configPath); err != nil {
		return crex.SystemError(description, "the local virtual machine could not be created").
			Recovery("Retry after recreating the local virtual machine.").
			Cause(crex.Wrap(ErrHostCreate, err)).
			Err()
	}

	if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
		return crex.SystemError("cannot start local environment", "the local virtual machine failed to start").
			Recovery("Retry after recreating the local environment.").
			Cause(crex.Wrap(ErrHostStart, err)).
			Err()
	}

	return nil
}

// Downloads and caches the machine disk image from the registry.
//
// Pulls the pinned machine version and extracts it into the local registry
// cache. After this returns, [cachedMachineImagePath] should resolve. The
// image is built and published by the Crucible team. It's an Alpine image
// with containerd installed and used as the base for provisioning the VM.
func fetchMachineImage(ctx context.Context, registryURL *url.URL) error {
	const description = "cannot download machine image"
	if registryURL == nil {
		return crex.ProgrammingError(description, "the machine image registry URL is required").
			Err()
	}

	registry := registryURL.String()
	src, err := hub.NewSource(registry, machineNamespace)
	if err != nil {
		return crex.SystemError(description, "the registry source could not be initialized").
			Recovery("Check your network connection and try again.").
			Cause(err).
			Err()
	}

	id := reference.NewIdentifier(machineType, registry, machineNamespace, machineName)
	ref, err := reference.New(id, machineVersion, nil)
	if err != nil {
		return crex.SystemError(description, "the machine image reference is invalid").
			Recovery("If the problem persists, report it to the Crucible team.").
			Cause(err).
			Err()
	}

	if _, err := src.Pull(ctx, ref); err != nil {
		return crex.SystemError(description, "the machine image could not be retrieved from the registry").
			Recovery("Check your network connection and try again.").
			Cause(err).
			Err()
	}
	return nil
}

// Returns the local path of the cached machine disk image.
//
// Checks the expected cache location for the image file. If the file is not
// found, returns [ErrMachineImageMissing]. It's up to the caller to decide
// whether to attempt a fetch from the registry.
func cachedMachineImagePath() (string, error) {
	c, err := cache.Open()
	if err != nil {
		return "", err
	}
	defer c.Close()

	ok, err := c.HasExtracted(machineNamespace, machineName, machineVersion)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", crex.Newf(ErrMachineImageMissing, "%s/%s %s is not in the cache", machineNamespace, machineName, machineVersion)
	}
	dir, err := c.Extract(machineNamespace, machineName, machineVersion)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s%s", machineArch(), machineExtension))
	if _, err := os.Stat(path); err != nil {
		return "", crex.Newf(ErrMachineImageMissing, "expected at %s", path)
	}
	return path, nil
}

// Returns the disk image architecture identifier for the current platform
// using Lima's naming convention (x86_64, aarch64).
func machineArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return limaArchARM64
	default:
		return limaArchAMD64
	}
}

// Deletes the host VM and its disk images.
//
// Blocks until cleanup is complete. Returns [ErrHostNotCreated] if the VM
// does not exist.
func destroyHost(ctx context.Context) error {
	status := hostStatus(ctx)
	if status == StateNotProvisioned {
		return ErrHostNotCreated
	}

	if err := limactlRun(ctx, "delete", "--force", limaInstanceName); err != nil {
		return crex.SystemError("cannot destroy local environment", "the local virtual machine could not be deleted").
			Recovery("Try again, or stop the local environment first.").
			Cause(crex.Wrap(ErrHostDestroy, err)).
			Err()
	}

	return nil
}

// Runs a command inside the host VM.
//
// Executes via limactl shell. Streams stdout and stderr to the provided writers
// and returns the command's exit code.
func hostExec(ctx context.Context, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return limaGuestExec(ctx, stdout, stderr, command, args...)
}

// Returns the host address where cruxd is reachable for the named instance.
func cruxdAddress(_ context.Context, _ string) (string, error) {
	return net.JoinHostPort(cruxdHost, strconv.Itoa(speccompute.DefaultPort)), nil
}

// Queries the current state of the host VM.
//
// Maps the Lima instance status string to a local state. If the instance
// does not exist or cannot be reached, returns [StateNotProvisioned].
func hostStatus(ctx context.Context) State {
	switch limaInstanceStatus(ctx) {
	case limaStatusRunning:
		return StateRunning
	case limaStatusStopped:
		return StateStopped
	default:
		return StateNotProvisioned
	}
}
