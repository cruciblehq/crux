package compute

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/cruciblehq/crux/compute/lima"
	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/crux/machines"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Name of the local machine managed by crux.
const localMachineName = "crux"

// Returns the identifier of the local machine.
func DefaultLocalMachineID() string {
	return localMachineName
}

// [Provider] for a single Lima VM on this computer.
//
// Adds crux policy to the [lima] driver: machine enrollment, readiness
// checks, and sessions with cruxd. The machine identifier is always
// [localMachineName].
type localProvider struct{}

// Uploads a disk image to the local provider.
//
// Writes r to a temporary file and passes the path to the Lima driver. The
// file stays until [localProvider.RemoveImage], because Provision reads it.
func (p localProvider) UploadImage(ctx context.Context, r io.Reader) (string, error) {
	const description = "cannot stage machine image"
	const recoveryDiskSpace = "Free up disk space, then try again."

	// Local provisioning consumes an image path during a later call to
	// Provision, so this path must remain valid after Upload returns.
	if f, ok := r.(*os.File); ok {
		return lima.UploadImage(ctx, f.Name())
	}

	f, err := createTemp("upload-*.img")
	if err != nil {
		return "", crex.SystemError(description, "failed to create a temporary file for the image upload").
			Recovery(recoveryDiskSpace).
			Cause(err).
			Err()
	}
	name := f.Name()
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return "", crex.SystemError(description, "failed to write the image to a temporary file").
			Recovery(recoveryDiskSpace).
			Cause(err).
			Err()
	}
	if err := f.Close(); err != nil {
		return "", crex.SystemError(description, "failed to finalize the temporary image file").
			Recovery(recoveryDiskSpace).
			Cause(err).
			Err()
	}
	return lima.UploadImage(ctx, name)
}

// Removes a previously uploaded disk image.
//
// Deletes the image file if it resides in the local provider's temporary
// staging directory. Images outside the staging directory are not removed
// because they are owned by the caller.
func (p localProvider) RemoveImage(ctx context.Context, id string) error {
	return lima.RemoveImageIfStaged(ctx, id, dirs.Temp())
}

// Provisions a new machine from a previously uploaded image.
//
// The machine boots with enrollment data for a new machine record, and this
// blocks until crux has enrolled it and cruxd reports healthy.
func (p localProvider) Provision(ctx context.Context, img string, opts ProvisionOptions) (string, error) {
	id := localMachineName
	enrollment, err := newMachine(id)
	if err != nil {
		return "", err
	}
	if err := lima.Provision(ctx, img, enrollment); err != nil {
		return "", limaError(err)
	}
	if err := p.waitReady(ctx, id); err != nil {
		return "", err
	}
	return id, nil
}

// Tears down the machine and removes all associated persistent state.
func (p localProvider) Deprovision(ctx context.Context, id string) error {
	if err := lima.Deprovision(ctx); err != nil {
		return err
	}
	return machines.Remove(dirs.Machines(), id)
}

// Starts the machine and blocks until it is reachable.
func (p localProvider) Start(ctx context.Context, id string) error {
	if err := lima.Start(ctx); err != nil {
		return limaError(err)
	}
	return p.waitReady(ctx, id)
}

// Stops the machine without removing its persistent state.
func (p localProvider) Stop(ctx context.Context, id string) error {
	return limaError(lima.Stop(ctx))
}

// Returns the current lifecycle state of the machine.
//
// A running VM counts as running only when cruxd answers its health check
// over mutual TLS.
func (p localProvider) Status(ctx context.Context, id string) (State, error) {
	state, err := lima.Status(ctx)
	if err != nil {
		return 0, err
	}
	if state != lima.StateRunning {
		return limaStateToState(state), nil
	}
	if machineHealthy(ctx, id, lima.CruxdAddress()) != nil {
		return StateStopped, nil
	}
	return StateRunning, nil
}

// Returns the identifiers of all Lima instances.
func (p localProvider) List(ctx context.Context) ([]string, error) {
	return lima.List(ctx)
}

// Runs a command on the machine outside any container.
func (p localProvider) Exec(ctx context.Context, id string, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return lima.Exec(ctx, stdout, stderr, command, args...)
}

// Opens a session with cruxd on the machine over mutual TLS.
func (p localProvider) Connect(ctx context.Context, id string) (Machine, error) {
	return connectMachine(ctx, id, lima.CruxdAddress())
}

// Blocks until the machine is enrolled and cruxd reports healthy.
func (p localProvider) waitReady(ctx context.Context, id string) error {
	return waitForMachine(ctx, id, lima.CruxdAddress())
}

// Tags Lima lifecycle errors with the matching compute error.
func limaError(err error) error {
	switch {
	case errors.Is(err, lima.ErrMachineNotCreated):
		return crex.Tag(err, ErrMachineNotProvisioned)
	case errors.Is(err, lima.ErrMachineAlreadyProvisioned):
		return crex.Tag(err, ErrMachineAlreadyProvisioned)
	case errors.Is(err, lima.ErrMachineNotRunning):
		return crex.Tag(err, ErrMachineNotRunning)
	default:
		return err
	}
}

// Converts a [lima.State] value to a [State].
func limaStateToState(s lima.State) State {
	switch s {
	case lima.StateRunning:
		return StateRunning
	case lima.StateStopped:
		return StateStopped
	default:
		return StateNotProvisioned
	}
}

// Creates a temporary file for local compute staging.
func createTemp(pattern string) (*os.File, error) {
	dir := dirs.Temp()
	if err := os.MkdirAll(dir, file.OwnerDirMode); err != nil {
		return nil, err
	}
	return os.CreateTemp(dir, pattern)
}
