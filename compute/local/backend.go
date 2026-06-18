package local

import (
	"context"
	"io"

	"github.com/cruciblehq/crux/affordance/kernel"
)

// The local compute backend.
//
// All lifecycle methods call the platform-specific implementation defined in
// the corresponding build-tagged source files.
type Backend struct{}

// Returns a new local backend.
func NewBackend() *Backend {
	return &Backend{}
}

// Ensures the machine disk image is available locally and returns its path.
//
// Resolves the default machine image from the local cache and downloads it
// from the Crucible registry if missing. The returned path can be passed
// directly to [provider.Backend.UploadImage].
func EnsureMachineImage(ctx context.Context) (string, error) {
	return ensureMachineImage(ctx)
}

// Uploads a disk image to the local provider, validating it is accessible.
func (b *Backend) UploadImage(ctx context.Context, path string) (string, error) {
	return uploadImage(ctx, path)
}

// Provisions a compute host instance from a previously uploaded image.
func (b *Backend) Provision(ctx context.Context, name, imageID string, kernelSpec kernel.Spec) error {
	return provision(ctx, name, imageID, kernelSpec)
}

// Tears down the instance and removes all state.
func (b *Backend) Deprovision(ctx context.Context, name string) error {
	return deprovision(ctx, name)
}

// Starts a previously provisioned instance.
func (b *Backend) Start(ctx context.Context, name string) error {
	return start(ctx, name)
}

// Stops a running instance.
func (b *Backend) Stop(ctx context.Context, name string) error {
	return stop(ctx, name)
}

// Returns the current state of the given instance.
func (b *Backend) Status(ctx context.Context, name string) (State, error) {
	return status(ctx, name)
}

// Lists all instances managed by the local provider.
func (b *Backend) List(ctx context.Context) ([]string, error) {
	return list(ctx)
}

// Sends a tar archive to the named instance and applies it to the host filesystem.
func (b *Backend) Copy(ctx context.Context, name string, r io.Reader) error {
	return copyArchive(ctx, name, r)
}

// Runs a command on the given instance, streaming output to stdout and stderr.
func (b *Backend) Run(ctx context.Context, name string, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return execute(ctx, name, stdout, stderr, command, args...)
}

// Returns the containerd socket path for the given instance.
func (b *Backend) ContainerdSocket(ctx context.Context, name string) (string, error) {
	return containerdSocket(ctx, name)
}
