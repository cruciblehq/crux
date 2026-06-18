//go:build darwin || linux

package local

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cruciblehq/crux/affordance/kernel"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/registry"
)

// Machine image coordinates.
const (
	machineNamespace   = "crucible"                       // Registry namespace for the machine image.
	machineName        = "machine-default"                // Registry resource name for the machine image.
	machineVersion     = "0.1.0"                          // Pinned machine image version.
	machineRegistryURL = "http://hub.cruciblehq.xyz:8080" // Registry URL for the machine image.
	machineExtension   = ".qcow2"                         // Disk image file extension.
)

// Containerd readiness polling.
const (
	containerdReadyTimeout = 15 * time.Minute // Maximum time to wait for containerd to start.
	containerdPollInterval = 2 * time.Second  // Interval between containerd readiness polls.
)

// Ensures the default machine disk image is available in the local cache.
//
// Downloads from the Crucible registry if not already cached. Returns the
// local filesystem path to the image file.
func ensureMachineImage(ctx context.Context) (string, error) {
	path, err := cachedMachineImagePath()
	if errors.Is(err, ErrMachineImageMissing) {
		slog.Info("machine image not in cache, downloading...",
			"namespace", machineNamespace,
			"name", machineName,
			"version", machineVersion,
		)
		if err := fetchMachineImage(ctx); err != nil {
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
// When this function is called, other providers copy the image file to remote
// storage and return an image ID. The local provider uses the image directly
// from the local filesystem, so this function only verifies the file exists
// and is accessible. The returned path is used as the image ID passed to Lima
// during provisioning.
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
// [uploadImage]. Returns [ErrHostAlreadyProvisioned] if an instance already
// exists. The VM is created and started.
func provision(ctx context.Context, _, imageID string, kernelSpec kernel.Spec) error {
	if err := ensureLima(ctx); err != nil {
		return err
	}
	if hostStatus(ctx) != StateNotProvisioned {
		return ErrHostAlreadyProvisioned
	}
	return createAndStartHost(ctx, imageID, kernelSpec)
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

// Queries the current state of a compute instance.
//
// The state is determined by probing the Lima VM and the containerd socket
// inside it, the returned state being the least-healthy of the two.
func status(ctx context.Context, _ string) (State, error) {
	rtState := hostStatus(ctx)

	if rtState == StateNotProvisioned {
		return StateNotProvisioned, nil
	}
	if rtState != StateRunning {
		return StateStopped, nil
	}

	if !isContainerdReady(ctx) {
		return StateStopped, nil
	}
	return StateRunning, nil
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

// Blocks until containerd is accepting connections.
//
// Polls via gRPC Health.Check every two seconds. Returns an error if the
// context is cancelled or the fifteen-minute deadline is exceeded.
func waitForContainerd(ctx context.Context) error {
	deadline := time.Now().Add(containerdReadyTimeout)
	for {
		if isContainerdReady(ctx) {
			return nil
		}
		if time.Now().After(deadline) {
			return crex.SystemErrorf("local runtime not ready", "containerd did not become ready within %s", containerdReadyTimeout).
				Recovery("Recreate the local runtime and retry.").
				Cause(ErrHostStart).
				Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(containerdPollInterval):
		}
	}
}

// Whether containerd is ready to accept gRPC connections.
//
// Issues a gRPC Health.Check RPC against the forwarded containerd socket.
// Returns true only when containerd responds with SERVING status.
func isContainerdReady(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		"unix://"+files.ContainerdSocket(limaInstanceName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false
	}
	defer conn.Close()

	resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	return err == nil && resp.Status == grpc_health_v1.HealthCheckResponse_SERVING
}

// Creates the Lima instance from the given disk image and starts it.
//
// Lima 2.x create does not auto-start; an explicit start call is required.
// imagePath is the local filesystem path to a QCOW2 disk image. kernelSpec
// configures kernel-level requirements; a zero value applies no additional
// requirements.
func createAndStartHost(ctx context.Context, imagePath string, kernelSpec kernel.Spec) error {
	const description = "cannot create local environment"

	configPath, err := generateLimaConfig(imagePath, kernelSpec)
	if err != nil {
		return crex.SystemError(description, "failed to generate the virtual machine configuration").
			Recovery("Regenerate the local virtual machine configuration and retry.").
			Cause(crex.Wrap(ErrHostCreate, err)).
			Err()
	}

	if err := limactlRunNoTTY(ctx, "create", "--name="+limaInstanceName, configPath); err != nil {
		return crex.SystemError(description, "the local virtual machine could not be created").
			Recovery("Retry after recreating the local virtual machine.").
			Cause(crex.Wrap(ErrHostCreate, err)).
			Err()
	}

	if err := removeInstanceSocket(); err != nil {
		return crex.SystemError(description, "failed to reset the local runtime socket").
			Recovery("Reset the local runtime state and retry.").
			Cause(crex.Wrap(ErrHostCreate, err)).
			Err()
	}

	if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
		return crex.SystemError("cannot start local environment", "the local virtual machine failed to start").
			Recovery("Retry after recreating the local environment.").
			Cause(crex.Wrap(ErrHostStart, err)).
			Err()
	}

	if err := waitForContainerd(ctx); err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	return nil
}

// Downloads and caches the machine disk image from the Crucible registry.
//
// Pulls the pinned machine version and extracts it into the local registry
// cache. After this returns, [cachedMachineImagePath] should resolve. The
// image is built and published by the Crucible team. It's an Alpine image
// with containerd installed and used as the base for provisioning the VM.
func fetchMachineImage(ctx context.Context) error {
	const description = "cannot download machine image"

	src, err := registry.NewSource(machineRegistryURL, machineNamespace)
	if err != nil {
		return crex.SystemError(description, "the registry source could not be initialized").
			Recovery("Check your network connection and try again.").
			Cause(err).
			Err()
	}

	id := reference.NewIdentifier(machineName, machineRegistryURL, machineNamespace, machineName)
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
	arch := machineArch()
	path := filepath.Join(
		files.RegistryExtractedVersionDir(machineNamespace, machineName, machineVersion),
		arch+machineExtension,
	)
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

// Removes the host-side containerd socket directory for the Lima instance.
//
// Lima binds a Unix listener to the host socket path at startup. If a previous
// hostagent did not exit cleanly, the socket file is left on disk and the next
// start fails with "address already in use". Removing the directory before
// start ensures Lima can always bind a fresh listener.
func removeInstanceSocket() error {
	socketDir := filepath.Dir(files.ContainerdSocket(limaInstanceName))
	if err := os.RemoveAll(socketDir); err != nil {
		return crex.Wrap(ErrHostStart, err)
	}
	return nil
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
			Recovery("Try again, or stop the local environment first with 'crux local stop'.").
			Cause(crex.Wrap(ErrHostDestroy, err)).
			Err()
	}

	// limactl delete does not stop the VM first, so the forwarded socket may
	// be left on disk. Remove the instance socket directory so a subsequent
	// start can bind a fresh listener.
	if err := removeInstanceSocket(); err != nil {
		return crex.SystemError("cannot destroy local environment", "failed to remove the local runtime socket").
			Recoveryf("Make sure you can delete %s, then try again.", filepath.Dir(files.ContainerdSocket(limaInstanceName))).
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

// Returns the containerd socket path for the named instance.
func containerdSocket(_ context.Context, name string) (string, error) {
	return files.ContainerdSocket(name), nil
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
