package compute

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Name of the local VM instance managed by crux.
const localInstanceName = "crux"

// Returns the default instance identifier used by the local backend.
func DefaultLocalInstanceID() string {
	return localInstanceName
}

// Permission mode for the temp directory, reachable only by the current user.
const tempDirMode os.FileMode = 0o700

// Local port of the compute backend.
type BackendLocal struct {
	compute Compute // Compute instance manager.
	storage Storage // Persistent volume manager.
	network Network // Network topology and firewall manager.
}

// Shim that adapts the local platform implementation to the [Compute] interface.
type computeLocalShim struct {
	local *local.Backend // Local backend implementation.
}

// Returns a [Backend] backed by the local platform implementation.
func newBackendLocal() Backend {
	compute := &computeLocalShim{local: local.NewBackend()}
	bl := &BackendLocal{
		compute: compute,
		storage: newStorageLocal(filepath.Join(localDataDir(), "storage")),
	}
	bl.network = newNetworkLocal(compute.vmExec)
	return bl
}

// Adapts the local backend's Run method to the [local.ExecFunc] signature.
func (s *computeLocalShim) vmExec(ctx context.Context, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return s.local.Run(ctx, localInstanceName, stdout, stderr, command, args...)
}

// Returns the compute instance manager.
func (bl *BackendLocal) Compute() Compute {
	return bl.compute
}

// Returns the persistent volume manager.
func (bl *BackendLocal) Storage() Storage {
	return bl.storage
}

// Returns the network topology and firewall manager.
func (bl *BackendLocal) Network() Network {
	return bl.network
}

// Uploads a disk image to the local provider.
//
// Writes r to a temporary file and passes the path to the underlying local
// backend implementation. The temporary file is removed after upload.
func (s *computeLocalShim) UploadImage(ctx context.Context, r io.Reader) (string, error) {
	const description = "cannot stage machine image"
	const recoveryDiskSpace = "Free up disk space, then try again."

	// Local provisioning consumes an image path during a later call to
	// Provision, so this path must remain valid after Upload returns.
	if f, ok := r.(*os.File); ok {
		return s.local.UploadImage(ctx, f.Name())
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
	return s.local.UploadImage(ctx, name)
}

// Removes a previously uploaded disk image.
//
// Deletes the image file if it resides in the local provider's temporary
// staging directory. Images outside the staging directory are not removed
// because they are owned by the caller.
func (s *computeLocalShim) RemoveImage(ctx context.Context, id string) error {
	return local.RemoveImageIfStaged(ctx, id, tempDir())
}

// Provisions a new instance from a previously uploaded image.
func (s *computeLocalShim) Provision(ctx context.Context, img string, opts ComputeOptions) (string, error) {
	id := localInstanceName
	if err := s.local.Provision(ctx, id, img); err != nil {
		return "", err
	}
	return id, nil
}

// Tears down the instance and removes all associated persistent state.
func (s *computeLocalShim) Deprovision(ctx context.Context, id string) error {
	return s.local.Deprovision(ctx, id)
}

// Starts the instance and blocks until it is reachable.
func (s *computeLocalShim) Start(ctx context.Context, id string) error {
	return s.local.Start(ctx, id)
}

// Stops the instance without removing its persistent state.
func (s *computeLocalShim) Stop(ctx context.Context, id string) error {
	return s.local.Stop(ctx, id)
}

// Returns the current lifecycle state of the instance.
func (s *computeLocalShim) Status(ctx context.Context, id string) (State, error) {
	state, err := s.local.Status(ctx, id)
	if err != nil {
		return 0, err
	}
	return localStateToState(state), nil
}

// Returns the identifiers of all instances known to the local backend.
func (s *computeLocalShim) List(ctx context.Context) ([]string, error) {
	return s.local.List(ctx)
}

// Runs a command on the instance outside any container.
func (s *computeLocalShim) Exec(ctx context.Context, id string, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return s.local.Run(ctx, id, stdout, stderr, command, args...)
}

// Sends a tar archive to the instance and applies it to its filesystem.
func (s *computeLocalShim) Copy(ctx context.Context, id string, r io.Reader) error {
	return s.local.Copy(ctx, id, r)
}

// Opens a client connection to the container runtime on the instance.
func (s *computeLocalShim) Connect(ctx context.Context, id string) (*Client, error) {
	socketPath, err := s.local.ContainerdSocket(ctx, id)
	if err != nil {
		return nil, err
	}
	return newClient(socketPath, localInstanceName)
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

// Path to the directory for temporary local compute files.
func tempDir() string {
	name := localInstanceName
	if uid := os.Getuid(); uid >= 0 {
		name = fmt.Sprintf("%s-%s", name, strconv.Itoa(uid))
	}
	return filepath.Join(os.TempDir(), name)
}

// Path to the persistent data directory for the local provider.
func localDataDir() string {
	name := localInstanceName
	if uid := os.Getuid(); uid >= 0 {
		name = fmt.Sprintf("%s-%s", name, strconv.Itoa(uid))
	}
	return filepath.Join(os.TempDir(), name, "data")
}

// Creates a temporary file for local compute staging.
func createTemp(pattern string) (*os.File, error) {
	dir := tempDir()
	if err := os.MkdirAll(dir, tempDirMode); err != nil {
		return nil, err
	}
	if err := verifyTempDirPermissions(dir); err != nil {
		return nil, err
	}
	return os.CreateTemp(dir, pattern)
}

// Validates a temp directory and repairs permissions when they are too open.
//
// dir must be a real directory rather than a symlink. If the directory exists
// but is writable by group or other users, the permissions are tightened to
// the private temporary-directory mode before the caller proceeds.
func verifyTempDirPermissions(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return crex.Newf(file.ErrUnsafeTempDir, "%q is a symlink or not a directory", dir)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return os.Chmod(dir, tempDirMode)
	}
	return nil
}
