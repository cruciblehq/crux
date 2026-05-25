//go:build darwin || linux

package local

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/source"
)

const (
	// Crucible machine image registry coordinates.
	machineNamespace   = "crucible"                       // Registry namespace for the machine image.
	machineName        = "machine"                        // Registry resource name for the machine image.
	machineVersion     = "0.1.8"                          // Pinned machine image version.
	machineRegistryURL = "http://hub.cruciblehq.xyz:8080" // Registry URL for the machine image.
	machineExtension   = ".qcow2"                         // Disk image file extension.

	// limaNoTTY disables terminal interaction for unattended limactl commands.
	limaNoTTY = "--tty=false"
)

// Ensures the host VM is running, provisioning it if necessary.
//
// If the VM does not exist, it is created from the cached machine image and
// started. If the VM exists but is stopped, it is resumed. If already running,
// this is a no-op.
func ensureHostRunning(ctx context.Context, name string) error {
	if err := ensureLima(ctx); err != nil {
		return err
	}

	status, err := hostStatus(ctx)
	if err != nil {
		return err
	}

	switch status {
	case provider.StateRunning:
		return nil

	case provider.StateStopped:
		if err := removeInstanceSocket(); err != nil {
			return err
		}
		if err := limaRun(ctx, "start", limaNoTTY, limaInstanceName); err != nil {
			return crex.Wrap(ErrHostStart, err)
		}
		if err := waitForContainerd(ctx); err != nil {
			return crex.Wrap(ErrHostStart, err)
		}
		return nil

	case provider.StateNotProvisioned:
		return createAndStartHost(ctx, name)

	default:
		return crex.Wrapf(ErrHostStart, "unexpected VM state: %s", status)
	}
}

// Creates the Lima instance from the cached machine image and starts it.
//
// Lima 2.x create does not auto-start; an explicit start call is required.
func createAndStartHost(ctx context.Context, name string) error {
	imagePath, err := cachedMachineImagePath()
	if errors.Is(err, ErrMachineImageMissing) {
		if err := fetchMachineImage(ctx); err != nil {
			return crex.Wrap(ErrHostCreate, err)
		}
		imagePath, err = cachedMachineImagePath()
	}
	if err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	configPath, err := generateLimaConfig(name, imagePath)
	if err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	if err := limaRun(ctx, "create", "--name="+limaInstanceName, limaNoTTY, configPath); err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	if err := removeInstanceSocket(); err != nil {
		return crex.Wrap(ErrHostCreate, err)
	}

	if err := limaRun(ctx, "start", limaNoTTY, limaInstanceName); err != nil {
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
// will resolve successfully.
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

	ref, err := reference.Parse(machineNamespace+"/"+machineName+" "+machineVersion, machineName)
	if err != nil {
		return err
	}
	ref = ref.WithDefaults(machineRegistryURL, machineNamespace)

	if _, err := src.Pull(ctx, ref); err != nil {
		return err
	}
	return nil
}

// Returns the local path of the cached machine disk image for the current
// architecture.
//
// Returns [ErrMachineImageMissing] if the image has not been downloaded yet.
func cachedMachineImagePath() (string, error) {
	arch := machineArch()
	path := filepath.Join(
		paths.RegistryCacheDir(), "extracted",
		machineNamespace, machineName, machineVersion,
		arch+machineExtension,
	)
	if _, err := os.Stat(path); err != nil {
		return "", crex.Wrapf(ErrMachineImageMissing, "expected at %s", path)
	}
	return path, nil
}

// Returns the disk image architecture identifier for the current platform.
//
// Uses Lima's architecture naming convention (aarch64, x86_64).
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
// Lima binds a Unix listener to the host socket path at startup. If a
// previous hostagent exited uncleanly, the socket file is left on disk and
// the next start fails with "address already in use". Removing the directory
// before each start ensures Lima can always bind a fresh listener.
func removeInstanceSocket() error {
	socketDir := filepath.Dir(paths.ContainerdSocket(limaInstanceName))
	if err := os.RemoveAll(socketDir); err != nil {
		return crex.Wrap(ErrHostStart, err)
	}
	return nil
}

// Deletes the host VM and its disk images.
//
// Blocks until cleanup is complete.
func destroyHost(ctx context.Context) error {
	status, err := hostStatus(ctx)
	if err != nil {
		return err
	}
	if status == provider.StateNotProvisioned {
		return ErrHostNotCreated
	}

	if err := limaRun(ctx, "delete", "--force", limaInstanceName); err != nil {
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

// Runs a command inside the host VM and captures its output.
func hostExec(ctx context.Context, command string, args ...string) (*provider.ExecResult, error) {
	return limaExec(ctx, command, args...)
}

// Queries the current state of the host VM.
//
// Maps the Lima instance status to a provider state.
func hostStatus(ctx context.Context) (provider.State, error) {
	switch limaInstanceStatus(ctx) {
	case limaStatusRunning:
		return provider.StateRunning, nil
	case limaStatusStopped:
		return provider.StateStopped, nil
	default:
		return provider.StateNotProvisioned, nil
	}
}
