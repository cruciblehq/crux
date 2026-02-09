package runtime

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/errdefs"
	"github.com/cruciblehq/crux/kit/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"
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
// The registry is the containerd namespace (typically the Crucible registry
// authority). The id is the container identifier within that namespace.
// This does not create or start anything in the runtime.
func NewContainer(registry, id string) *Container {
	return &Container{registry: registry, id: id}
}

// Sequence counter for generating unique exec process identifiers.
var execSeq uint64

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
// The command runs within the container's task as an exec process. The
// process inherits the container's OCI spec (environment, working
// directory, capabilities). The container must be running.
func (c *Container) Exec(ctx context.Context, command string, args ...string) (*ExecResult, error) {
	client, err := newContainerdClient(c.registry)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}
	defer client.Close()

	ctr, err := client.LoadContainer(ctx, c.id)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	task, err := ctr.Task(ctx, nil)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	pspec, err := execSpec(ctx, ctr, command, args...)
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	return runExec(ctx, task, pspec)
}

// Builds a process spec for an exec by cloning the container's OCI
// process configuration and overriding the arguments.
func execSpec(ctx context.Context, ctr containerd.Container, command string, args ...string) (*specs.Process, error) {
	spec, err := ctr.Spec(ctx)
	if err != nil {
		return nil, err
	}

	pspec := *spec.Process
	pspec.Terminal = false
	pspec.Args = append([]string{command}, args...)
	return &pspec, nil
}

// Creates an exec process, runs it, and collects its output.
func runExec(ctx context.Context, task containerd.Task, pspec *specs.Process) (*ExecResult, error) {
	execID := fmt.Sprintf("exec-%d", atomic.AddUint64(&execSeq, 1))

	var stdout, stderr bytes.Buffer
	process, err := task.Exec(ctx, execID, pspec, cio.NewCreator(
		cio.WithStreams(nil, &stdout, &stderr),
	))
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	statusC, err := process.Wait(ctx)
	if err != nil {
		process.Delete(ctx)
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	if err := process.Start(ctx); err != nil {
		process.Delete(ctx)
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	exitStatus := <-statusC
	process.Delete(ctx)

	code, _, err := exitStatus.Result()
	if err != nil {
		return nil, crex.Wrap(ErrContainerExec, err)
	}

	return &ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: int(code),
	}, nil
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
