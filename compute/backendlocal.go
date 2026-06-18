package compute

import (
	"context"
	"io"
	"os"

	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
)

// Name of the local VM instance managed by crux.
const LocalInstance = files.DefaultClientName

// Local port of the compute backend.
//
// This shim translates the local package's platform-specific implementation to
// the compute package's provider-agnostic interface. The local package has no
// dependency on compute types, so this shim bridges the two.
type BackendLocal struct {
	local *local.Backend // Local backend implementation.
}

// Returns a [Backend] backed by the local platform implementation.
func newBackendLocal() Backend {
	return &BackendLocal{local: local.NewBackend()}
}

// Uploads a disk image to the local provider.
//
// Writes r to a temporary file and passes the path to the underlying local
// backend implementation. The temporary file is removed after upload.
func (bl *BackendLocal) Upload(ctx context.Context, r io.Reader) (string, error) {
	const description = "cannot stage machine image"
	const recoveryDiskSpace = "Free up disk space, then try again."

	// Local provisioning consumes an image path during a later call to Provision,
	// so this path must remain valid after Upload returns.
	if f, ok := r.(*os.File); ok {
		return bl.local.UploadImage(ctx, f.Name())
	}

	f, err := files.CreateTemp("upload-*.img")
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
	return bl.local.UploadImage(ctx, name)
}

// Provisions a new host from a previously uploaded image.
func (bl *BackendLocal) Provision(ctx context.Context, img, name string, opts Options) error {
	return bl.local.Provision(ctx, name, img, opts.Kernel)
}

// Tears down the named host and removes all associated persistent state.
func (bl *BackendLocal) Deprovision(ctx context.Context, name string) error {
	return bl.local.Deprovision(ctx, name)
}

// Starts the named host and blocks until it is reachable.
func (bl *BackendLocal) Start(ctx context.Context, name string) error {
	return bl.local.Start(ctx, name)
}

// Stops the named host without removing its persistent state.
func (bl *BackendLocal) Stop(ctx context.Context, name string) error {
	return bl.local.Stop(ctx, name)
}

// Returns the current lifecycle state of the named host.
func (bl *BackendLocal) Status(ctx context.Context, name string) (State, error) {
	s, err := bl.local.Status(ctx, name)
	if err != nil {
		return 0, err
	}
	return localStateToState(s), nil
}

// Returns the names of all hosts known to the local backend.
func (bl *BackendLocal) List(ctx context.Context) ([]string, error) {
	return bl.local.List(ctx)
}

// Runs a command on the named host outside any container.
func (bl *BackendLocal) Exec(ctx context.Context, name string, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return bl.local.Run(ctx, name, stdout, stderr, command, args...)
}

// Sends a tar archive to the named host and applies it to the host filesystem.
func (bl *BackendLocal) Copy(ctx context.Context, name string, r io.Reader) error {
	return bl.local.Copy(ctx, name, r)
}

// Opens a client connection to the container runtime on the named host.
func (bl *BackendLocal) Connect(ctx context.Context, name string) (*Client, error) {
	socketPath, err := bl.local.ContainerdSocket(ctx, name)
	if err != nil {
		return nil, err
	}
	return newClient(socketPath, files.DefaultClientName)
}

// Converts a [local.State] value to a [State].
func localStateToState(s local.State) State {
	switch s {
	case local.StateRunning:
		return StateRunning
	case local.StateStopped:
		return StateStopped
	default:
		return StateNotProvisioned
	}
}
