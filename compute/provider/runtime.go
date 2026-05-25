package provider

import (
	"context"
	"io"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Interface for container-level operations on a provisioned compute instance.
//
// A ContainerRuntime connects to the containerd daemon socket exposed by a
// running [Backend] and manages the lifecycle of individual containers. The
// The security policy applied to each container comes from the [ExecConfig.Security]
// field, which is the compiled output of the runtime subsystems. Infrastructure
// fields (rootfs path, cgroup path, standard mounts) are managed by the
// runtime implementation using the snapshot created from the image.
type ContainerRuntime interface {

	// Loads an OCI image archive into the runtime's image store.
	//
	// The reader must contain a valid OCI image tar (as produced by
	// crane.Save or similar). Returns the image name for use with [Run].
	Import(ctx context.Context, r io.Reader) (string, error)

	// Creates, starts, and waits for a container, then removes it.
	//
	// Blocks until the container process exits or the context is cancelled.
	// Returns the container's exit code. The container and its snapshot are
	// deleted before returning regardless of outcome.
	Run(ctx context.Context, cfg *ExecConfig) (int, error)

	// Releases the connection to the containerd daemon.
	Close() error
}

// Configuration for running a single container.
type ExecConfig struct {

	// Unique identifier for the container within the runtime namespace.
	ID string

	// Image reference as returned by [ContainerRuntime.Import] or a
	// registry reference already known to the runtime's image store.
	Image string

	// Security and resource policy compiled from runtime subsystem grants.
	//
	// The implementation merges this spec with the image's own process
	// configuration. Infrastructure fields (Root.Path, CgroupsPath) are
	// set by the runtime; callers must not set them.
	Security *specs.Spec

	// Process command and arguments.
	//
	// Overrides the image's default entrypoint when non-empty.
	Args []string

	// Environment variables injected into the container process.
	//
	// Merged with the image's own environment; caller-supplied values win
	// on conflicts.
	Env []string

	// Working directory inside the container.
	//
	// Overrides the image's default when non-empty.
	WorkingDir string
}
