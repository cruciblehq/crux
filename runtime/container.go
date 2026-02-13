package runtime

import (
	"context"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/errdefs"
	"github.com/cruciblehq/crux/kit/crex"
)

// A container instance within the container runtime.
//
// Use [NewContainer] to construct a handle for an existing container, or
// [Image.Start] to create and start one.
type Container struct {
	registry string // Containerd namespace.
	id       string // Container identifier.
}

// Creates a [Container] handle for an existing container.
//
// The registry is the containerd namespace (the registry host authority).
// The id is the container identifier within that namespace.
// This does not create or start anything in the runtime.
func NewContainer(registry, id string) *Container {
	return &Container{registry: registry, id: id}
}

// Stops the container's task.
//
// The running task is killed and deleted. The container metadata is
// preserved. Stop is idempotent; calling it on an already-stopped
// container is not an error.
func (c *Container) Stop(ctx context.Context) error {
	client, err := newContainerdClient(c.registry)
	if err != nil {
		return crex.Wrap(ErrContainerStop, err)
	}
	defer client.Close()

	ctr, err := client.LoadContainer(ctx, c.id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return crex.Wrap(ErrContainerStop, err)
	}

	task, err := ctr.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return crex.Wrap(ErrContainerStop, err)
	}

	task.Kill(ctx, syscall.SIGKILL)
	if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil && !errdefs.IsNotFound(err) {
		return crex.Wrap(ErrContainerStop, err)
	}

	return nil
}

// Destroys the container.
//
// The task is killed and the container is removed from the runtime along
// with its snapshot. The image is not affected. After destruction the
// container cannot be restarted.
func (c *Container) Destroy(ctx context.Context) error {
	client, err := newContainerdClient(c.registry)
	if err != nil {
		return crex.Wrap(ErrContainerDestroy, err)
	}
	defer client.Close()

	ctr, loadErr := client.LoadContainer(ctx, c.id)
	if loadErr != nil {
		if errdefs.IsNotFound(loadErr) {
			return nil
		}
		return crex.Wrap(ErrContainerDestroy, loadErr)
	}

	if task, taskErr := ctr.Task(ctx, nil); taskErr == nil {
		task.Kill(ctx, syscall.SIGKILL)
		task.Delete(ctx, containerd.WithProcessKill)
	}

	if err := ctr.Delete(ctx, containerd.WithSnapshotCleanup); err != nil && !errdefs.IsNotFound(err) {
		return crex.Wrap(ErrContainerDestroy, err)
	}

	return nil
}

// Runs a command inside the container and captures its output.
//
// Equivalent to [Container.ExecWith] with zero-value options. The
// process inherits the container's OCI spec unchanged.
func (c *Container) Exec(ctx context.Context, command string, args ...string) (*ExecResult, error) {
	return c.ExecWith(ctx, ExecOptions{}, command, args...)
}

// Runs a command inside the container with custom options.
//
// The command runs within the container's task as an exec process.
// Options override the inherited OCI spec for environment and working
// directory. The container must be running.
func (c *Container) ExecWith(ctx context.Context, opts ExecOptions, command string, args ...string) (*ExecResult, error) {
	return containerExec(ctx, c.registry, c.id, opts, command, args...)
}

// Queries the current state of the container.
//
// Returns [StateRunning] if the task is running, [StateStopped] if the
// container exists but has no running task, or [StateNotCreated] if the
// container does not exist.
func (c *Container) Status(ctx context.Context) (State, error) {
	client, err := newContainerdClient(c.registry)
	if err != nil {
		return StateNotCreated, crex.Wrap(ErrContainerStatus, err)
	}
	defer client.Close()

	ctr, err := client.LoadContainer(ctx, c.id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return StateNotCreated, nil
		}
		return StateNotCreated, crex.Wrap(ErrContainerStatus, err)
	}

	task, err := ctr.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return StateStopped, nil
		}
		return StateNotCreated, crex.Wrap(ErrContainerStatus, err)
	}

	status, err := task.Status(ctx)
	if err != nil {
		return StateNotCreated, crex.Wrap(ErrContainerStatus, err)
	}

	switch status.Status {
	case containerd.Running:
		return StateRunning, nil
	default:
		return StateStopped, nil
	}
}
