package local

import (
	"context"
	"io"

	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/manifest"
)

// The local compute backend.
//
// All lifecycle methods call the platform-specific implementation defined in
// the corresponding build-tagged source files.
type backend struct{}

// Returns a new local [provider.Backend].
func NewBackend() provider.Backend {
	return &backend{}
}

// Ensures the machine disk image is available locally and returns its path.
//
// Downloads the image from the Crucible registry if it is not already cached.
// The returned path can be passed directly to [provider.Backend.UploadImage].
func EnsureMachineImage(ctx context.Context) (string, error) {
	return ensureMachineImage(ctx)
}

// Uploads a disk image to the local provider, validating it is accessible.
func (b *backend) UploadImage(ctx context.Context, path string) (string, error) {
	return uploadImage(ctx, path)
}

// Provisions a compute host instance from a previously uploaded image.
func (b *backend) Provision(ctx context.Context, name, imageID string, policy *manifest.ComputePolicy) error {
	return provision(ctx, name, imageID, policy)
}

// Tears down the instance and removes all state.
func (b *backend) Deprovision(ctx context.Context, name string) error {
	return deprovision(ctx, name)
}

// Starts a previously provisioned instance.
func (b *backend) Start(ctx context.Context, name string) error {
	return start(ctx, name)
}

// Stops a running instance.
func (b *backend) Stop(ctx context.Context, name string) error {
	return stop(ctx, name)
}

// Returns the current state of the given instance.
func (b *backend) Status(ctx context.Context, name string) (provider.State, error) {
	return status(ctx, name)
}

// Runs a command on the given instance, streaming output to stdout and stderr.
func (b *backend) Exec(ctx context.Context, name string, stdout, stderr io.Writer, command string, args ...string) (int, error) {
	return execute(ctx, name, stdout, stderr, command, args...)
}
