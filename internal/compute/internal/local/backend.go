package local

import (
	"context"

	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/resource"
)

// The local compute backend.
//
// All lifecycle methods delegate to platform-specific helpers defined in the
// corresponding build-tagged source files.
type backend struct{}

// Returns a new local [provider.Backend].
func NewBackend() provider.Backend {
	return &backend{}
}

// Provisions and starts a cruxd host instance.
func (b *backend) Provision(ctx context.Context, name string, source resource.Source) error {
	return provision(ctx, name, source)
}

// Tears down the cruxd instance and removes all state.
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

// Runs a command on the given instance and returns its output.
func (b *backend) Exec(ctx context.Context, name string, command string, args ...string) (*provider.ExecResult, error) {
	return execute(ctx, name, command, args...)
}

// Returns a [provider.Client] connected to the given instance.
func (b *backend) Client(_ context.Context, name string) (provider.Client, error) {
	return newClient(name)
}
