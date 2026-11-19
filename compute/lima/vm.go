//go:build darwin || linux

package lima

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/utils-go/crex"
)

// Host interface Lima forwards cruxd's port to.
const cruxdHost = "127.0.0.1"

// Checks that a machine disk image exists and returns its image ID.
//
// The local provider uses the image directly from the local filesystem, so this
// function only verifies the file exists and is accessible. The returned path
// is used as the image ID passed to Lima during provisioning. Other providers
// copy the image file to remote storage and return an image ID.
func UploadImage(_ context.Context, path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", crex.SystemError("cannot access machine image", "the machine image file is missing or unreadable").
			Recovery("Download the machine image again and retry.").
			Cause(crex.Wrap(ErrImageUpload, err)).
			Err()
	}
	return path, nil
}

// Creates and starts the machine from the given disk image.
//
// imageID is the local filesystem path to a QCOW2 disk image, as returned by
// [UploadImage]. enrollment is written into the VM so cruxd can enroll.
// Returns [ErrMachineAlreadyProvisioned] if the machine already exists. The VM is
// created and started.
func Provision(ctx context.Context, imageID string, enrollment []byte) error {
	if err := ensureLima(ctx); err != nil {
		return err
	}
	if machineStatus(ctx) != StateNotProvisioned {
		return ErrMachineAlreadyProvisioned
	}
	return createAndStartMachine(ctx, imageID, enrollment)
}

// Starts the VM.
//
// The VM must already be provisioned. If already running this is a no-op, and
// if it exists but is stopped, it is started. Returns [ErrMachineNotCreated] if
// the VM has not been provisioned.
func Start(ctx context.Context) error {
	state := machineStatus(ctx)
	if state == StateRunning {
		return nil
	}
	if state == StateNotProvisioned {
		return ErrMachineNotCreated
	}

	if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
		return crex.SystemError("cannot start local environment", "the local virtual machine failed to start").
			Recovery("Recreate the local virtual machine and retry.").
			Cause(crex.Wrap(ErrMachineStart, err)).
			Err()
	}
	return nil
}

// Stops the VM.
//
// Returns [ErrMachineNotRunning] if the VM is not currently running. limactl stop
// sends a graceful shutdown signal to the guest, giving containerd a chance to
// clean up before the VM halts.
func Stop(ctx context.Context) error {
	state := machineStatus(ctx)
	if state != StateRunning {
		return ErrMachineNotRunning
	}

	if err := limactlRun(ctx, "stop", limaInstanceName); err != nil {
		return crex.SystemError("cannot stop local environment", "the local virtual machine failed to stop").
			Recovery("Retry after recreating the local environment.").
			Cause(crex.Wrap(ErrMachineStop, err)).
			Err()
	}
	return nil
}

// Destroys the VM.
//
// After this, the VM no longer appears in limactl list and all resources it
// consumed are freed. If the VM exists, it is terminated and deleted along
// with its disk image. Returns [ErrMachineNotCreated] if the VM does not
// exist.
func Deprovision(ctx context.Context) error {
	return destroyMachine(ctx)
}

// Queries the current state of the VM.
//
// Reports only whether the VM runs. Whether cruxd is reachable is checked by
// the caller over mutual TLS.
func Status(ctx context.Context) (State, error) {
	return machineStatus(ctx), nil
}

// Runs a command inside the VM.
//
// Used for setup and teardown tasks that must run in the VM. The output is
// streamed to stdout and stderr as it is produced. Returns the exit code and
// a nil error when the process exits, whatever the exit code. Returns a
// non-nil error only if the command could not be started or the context was
// cancelled before the process completed.
func Exec(ctx context.Context, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return limaGuestExec(ctx, stdout, stderr, command, args...)
}

// Lists the Lima instances.
func List(ctx context.Context) ([]string, error) {
	return limaList(ctx)
}

// Creates the Lima instance from the given disk image and starts it.
//
// Lima 2.x create does not auto-start; an explicit start call is required.
// imagePath is the local filesystem path to a QCOW2 disk image. enrollment is
// written into the VM config. Returns once the VM runs, before cruxd is
// enrolled.
func createAndStartMachine(ctx context.Context, imagePath string, enrollment []byte) error {
	const description = "cannot create local environment"

	configPath, err := generateLimaConfig(imagePath, enrollment)
	if err != nil {
		return crex.SystemError(description, "failed to generate the virtual machine configuration").
			Recovery("Regenerate the local virtual machine configuration and retry.").
			Cause(crex.Wrap(ErrMachineCreate, err)).
			Err()
	}

	if err := limactlRunNoTTY(ctx, "create", fmt.Sprintf("--name=%s", limaInstanceName), configPath); err != nil {
		return crex.SystemError(description, "the local virtual machine could not be created").
			Recovery("Retry after recreating the local virtual machine.").
			Cause(crex.Wrap(ErrMachineCreate, err)).
			Err()
	}

	if err := limactlRunNoTTY(ctx, "start", limaInstanceName); err != nil {
		return crex.SystemError("cannot start local environment", "the local virtual machine failed to start").
			Recovery("Retry after recreating the local environment.").
			Cause(crex.Wrap(ErrMachineStart, err)).
			Err()
	}

	return nil
}

// Deletes the VM and its disk images.
//
// Blocks until cleanup is complete. Returns [ErrMachineNotCreated] if the VM
// does not exist.
func destroyMachine(ctx context.Context) error {
	status := machineStatus(ctx)
	if status == StateNotProvisioned {
		return ErrMachineNotCreated
	}

	if err := limactlRun(ctx, "delete", "--force", limaInstanceName); err != nil {
		return crex.SystemError("cannot destroy local environment", "the local virtual machine could not be deleted").
			Recovery("Try again, or stop the local environment first.").
			Cause(crex.Wrap(ErrMachineDestroy, err)).
			Err()
	}

	return nil
}

// Returns the address on this computer where cruxd in the VM is reachable.
func CruxdAddress() string {
	return net.JoinHostPort(cruxdHost, strconv.Itoa(agent.DefaultPort))
}

// Queries the current state of the VM.
//
// Maps the Lima instance status string to a [State]. If the instance
// does not exist or cannot be reached, returns [StateNotProvisioned].
func machineStatus(ctx context.Context) State {
	switch limaInstanceStatus(ctx) {
	case limaStatusRunning:
		return StateRunning
	case limaStatusStopped:
		return StateStopped
	default:
		return StateNotProvisioned
	}
}
