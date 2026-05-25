package compute

import (
	"github.com/cruciblehq/crux/compute/ctr"
	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/paths"
)

const (
	StateNotProvisioned = provider.StateNotProvisioned // Instance has not been provisioned.
	StateRunning        = provider.StateRunning        // Instance is running and reachable.
	StateStopped        = provider.StateStopped        // Instance exists but is not running.
)

// Interface for compute backend implementations.
type Backend = provider.Backend

// Lifecycle state of a compute instance.
type State = provider.State

// Output captured from a command executed on the instance's host.
type ExecResult = provider.ExecResult

// Returns a new [ExecResult].
var NewExecResult = provider.NewExecResult

// Interface for container-level operations on a provisioned compute instance.
type ContainerRuntime = provider.ContainerRuntime

// Configuration for running a single container.
type ExecConfig = provider.ExecConfig

// Interface for building OCI images via containerd primitives.
type ImageBuilder = provider.ImageBuilder

// Configuration for a single command executed inside a build container.
type RunConfig = provider.RunConfig

// Image configuration changes applied without adding a new layer.
type ConfigUpdate = provider.ConfigUpdate

// Connects to the containerd daemon for the given instance and returns a
// [ContainerRuntime].
//
// The instance must be provisioned and running before calling NewRuntime.
// The caller is responsible for calling Close on the returned runtime.
func NewRuntime(name string) (ContainerRuntime, error) {
	return ctr.New(paths.ContainerdSocket(name))
}

// Name of the local Lima VM instance managed by crux.
const LocalInstance = "crux"

// Connects to the containerd daemon for the given instance and returns an
// [ImageBuilder].
//
// The instance must be provisioned and running before calling NewImageBuilder.
// The caller is responsible for calling Close on the returned builder.
func NewImageBuilder(name string) (ImageBuilder, error) {
	return ctr.NewImageBuilder(paths.ContainerdSocket(name))
}

// Connects to the containerd daemon for the local compute instance and
// returns an [ImageBuilder].
//
// The local instance must be provisioned and running. The caller is
// responsible for calling Close on the returned builder.
func NewLocalImageBuilder() (ImageBuilder, error) {
	return NewImageBuilder(LocalInstance)
}
