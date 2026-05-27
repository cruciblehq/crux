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

	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/source"
)

const (

	// Registry details for the machine image.
	//
	// The image is built and published by the Crucible team. It's an Alpine
	// image with containerd installed. The local provider downloads the image
	// and uses it as the base image for provisioning the VM.
	machineNamespace   = "crucible"                       // Registry namespace for the machine image.
	machineName        = "machine"                        // Registry resource name for the machine image.
	machineVersion     = "0.1.8"                          // Pinned machine image version.
	machineRegistryURL = "http://hub.cruciblehq.xyz:8080" // Registry URL for the machine image.
	machineExtension   = ".qcow2"                         // Disk image file extension.
)

// Ensures the machine disk image is available in the local cache.
//
// Downloads from the Crucible registry if not already cached. Returns the
// local filesystem path to the image file.
func ensureMachineImage(ctx context.Context) (string, error) {
	path, err := cachedMachineImagePath()
	if errors.Is(err, ErrMachineImageMissing) {
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
		return "", crex.Wrap(ErrImageUpload, err)
	}
	return path, nil
}

// Provisions a compute instance from the given disk image.
//
// imageID is the local filesystem path to a QCOW2 disk image, as returned by
// [uploadImage]. The VM is created and started if it does not already exist.
// containerd runs as a system service inside the VM and starts automatically
// during boot. policy is applied to the VM configuration when creating a new
// instance; nil means no additional policy.
func provision(ctx context.Context, _, imageID string, policy *manifest.ComputePolicy) error {
	return ensureHostRunning(ctx, imageID, policy)
}

// Starts the VM.
//
// The VM must already be provisioned. If already running this is a no-op, and
// if it exists but is stopped, it is resumed. Returns [ErrHostNotCreated] if
// the VM has not been provisioned.
func start(ctx context.Context, _ string) error {
	state := hostStatus(ctx)
	if state == provider.StateRunning {
		return nil
	}
	if state == provider.StateNotProvisioned {
		return ErrHostNotCreated
	}

	if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
		return crex.Wrap(ErrHostStart, err)
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
	if state != provider.StateRunning {
		return ErrHostNotRunning
	}

	if err := limactlRun(ctx, "stop", limaInstanceName); err != nil {
		return crex.Wrap(ErrHostStop, err)
	}
	return nil
}

// Tears down the instance and destroys the VM.
//
// Returns nil if the VM does not exist (idempotent). If the VM exists, it is
// terminated and deleted along with its disk image. After this returns, the
// VM no longer appears in limactl list and all resources it consumed are freed.
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
func status(ctx context.Context, _ string) (provider.State, error) {
	rtState := hostStatus(ctx)

	if rtState == provider.StateNotProvisioned {
		return provider.StateNotProvisioned, nil
	}
	if rtState != provider.StateRunning {
		return provider.StateStopped, nil
	}

	if !isContainerdReady(ctx) {
		return provider.StateStopped, nil
	}
	return provider.StateRunning, nil
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

// Blocks until containerd is accepting connections.
//
// Polls via gRPC Health.Check every two seconds. Returns an error if the
// context is cancelled or the fifteen-minute deadline is exceeded.
func waitForContainerd(ctx context.Context) error {
	deadline := time.Now().Add(15 * time.Minute)
	for {
		if isContainerdReady(ctx) {
			return nil
		}
		if time.Now().After(deadline) {
			return crex.Wrapf(ErrHostStart, "timed out waiting for containerd to become ready")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
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
		"unix://"+paths.ContainerdSocket(limaInstanceName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false
	}
	defer conn.Close()

	resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	return err == nil && resp.Status == grpc_health_v1.HealthCheckResponse_SERVING
}

// Ensures the host VM is running, provisioning it if necessary.
//
// If the VM does not exist, it is created from imagePath and started. If the
// VM exists but is stopped, it is resumed. If already running, this is a
// no-op. policy is applied when creating a new VM and ignored when resuming
// an existing one.
func ensureHostRunning(ctx context.Context, imagePath string, policy *manifest.ComputePolicy) error {
	if err := ensureLima(ctx); err != nil {
		return err
	}

	status := hostStatus(ctx)

	switch status {
	case provider.StateRunning:
		return nil

	case provider.StateStopped:
		if err := removeInstanceSocket(); err != nil {
			return err
		}
		if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
			return crex.Wrap(ErrHostStart, err)
		}
		if err := waitForContainerd(ctx); err != nil {
			return crex.Wrap(ErrHostStart, err)
		}
		return nil

	case provider.StateNotProvisioned:
		return createAndStartHost(ctx, imagePath, policy)

	default:
		return crex.Wrapf(ErrHostStart, "unexpected VM state: %s", status)
	}
}

// Creates the Lima instance from the given disk image and starts it.
//
// Lima 2.x create does not auto-start; an explicit start call is required.
// policy configures VM-level security; nil applies no additional policy.
func createAndStartHost(ctx context.Context, imagePath string, policy *manifest.ComputePolicy) error {
	configPath, err := generateLimaConfig(imagePath, policy)
	if err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	if err := limactlRunNoTTY(ctx, "create", "--name="+limaInstanceName, configPath); err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	if err := removeInstanceSocket(); err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
		return crex.Wrap(ErrHostStart, err)
	}

	if err := waitForContainerd(ctx); err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	return nil
}

// Downloads and caches the machine disk image from the Crucible registry.
//
// Pulls the pinned machine version from the registry and extracts it into
// the local registry cache. After this returns, [cachedMachineImagePath]
// should resolve successfully.
func fetchMachineImage(ctx context.Context) error {
	slog.Info("machine image not in cache, downloading...",
		"namespace", machineNamespace,
		"name", machineName,
		"version", machineVersion,
	)

	src, err := source.NewSource(machineRegistryURL, machineNamespace)
	if err != nil {
		return err
	}

	id := reference.NewIdentifier(machineName, machineRegistryURL, machineNamespace, machineName)
	ref, err := reference.New(id, machineVersion, nil)
	if err != nil {
		return err
	}

	if _, err := src.Pull(ctx, ref); err != nil {
		return err
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
		paths.RegistryExtractedVersionDir(machineNamespace, machineName, machineVersion),
		arch+machineExtension,
	)
	if _, err := os.Stat(path); err != nil {
		return "", crex.Wrapf(ErrMachineImageMissing, "expected at %s", path)
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
	socketDir := filepath.Dir(paths.ContainerdSocket(limaInstanceName))
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
	if status == provider.StateNotProvisioned {
		return ErrHostNotCreated
	}

	if err := limactlRun(ctx, "delete", "--force", limaInstanceName); err != nil {
		return crex.Wrap(ErrHostDestroy, err)
	}

	// limactl delete does not stop the VM first, so the forwarded socket may
	// be left on disk. Remove the instance socket directory so a subsequent
	// start can bind a fresh listener.
	if err := removeInstanceSocket(); err != nil {
		return crex.Wrap(ErrHostDestroy, err)
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

// Queries the current state of the host VM.
//
// Maps the Lima instance status string to a provider state. If the instance
// does not exist or cannot be reached, returns [provider.StateNotProvisioned].
func hostStatus(ctx context.Context) provider.State {
	switch limaInstanceStatus(ctx) {
	case limaStatusRunning:
		return provider.StateRunning
	case limaStatusStopped:
		return provider.StateStopped
	default:
		return provider.StateNotProvisioned
	}
}
