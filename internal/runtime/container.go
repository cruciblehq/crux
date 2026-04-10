package runtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/containerd/errdefs"
	"github.com/cruciblehq/crex"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// A running build container backed by containerd.
type Container struct {
	client        *containerd.Client // Containerd client for managing the container.
	id            string             // Unique identifier for the container.
	platform      string             // OCI platform (e.g., "linux/amd64").
	hugepageSizes []string           // OCI hugepage sizes derived from the kernel selection (e.g., "2MB", "1GB").
}

// Queries the current state of the container.
//
// Returns [ContainerRunning] if the task is active, [ContainerStopped]
// if the container exists but has no running task, or [ContainerNotCreated]
// if the container does not exist.
func (c *Container) Status(ctx context.Context) (ContainerState, error) {
	ctr, err := c.client.LoadContainer(ctx, c.id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return ContainerNotCreated, nil
		}
		return "", crex.Wrap(ErrRuntime, err)
	}

	task, err := ctr.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return ContainerStopped, nil
		}
		return "", crex.Wrap(ErrRuntime, err)
	}

	status, err := task.Status(ctx)
	if err != nil {
		return "", crex.Wrap(ErrRuntime, err)
	}

	switch status.Status {
	case containerd.Running:
		return ContainerRunning, nil
	default:
		return ContainerStopped, nil
	}
}

// Stops the container's task.
//
// The running task is killed and deleted. The container metadata is preserved.
// Calling Stop on an already-stopped container is not an error.
func (c *Container) Stop(ctx context.Context) error {
	ctr, err := c.client.LoadContainer(ctx, c.id)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return crex.Wrap(ErrRuntime, err)
	}

	task, err := ctr.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return crex.Wrap(ErrRuntime, err)
	}

	task.Kill(ctx, syscall.SIGKILL)
	if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil && !errdefs.IsNotFound(err) {
		return crex.Wrap(ErrRuntime, err)
	}

	return nil
}

// Removes the container and its resources.
//
// The task is killed and the container is removed from containerd along with
// its snapshot. After destruction the handle is invalid.
func (c *Container) Destroy(ctx context.Context) {
	ctr, err := c.client.LoadContainer(ctx, c.id)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			slog.Error("failed to load container for destruction", "id", c.id, "error", err)
		}
		return
	}

	if task, err := ctr.Task(ctx, nil); err == nil {
		task.Kill(ctx, syscall.SIGKILL)
		task.Delete(ctx, containerd.WithProcessKill)
	}

	if err := ctr.Delete(ctx, containerd.WithSnapshotCleanup); err != nil && !errdefs.IsNotFound(err) {
		slog.Error("failed to delete container during destruction", "id", c.id, "error", err)
	}
}

// Starts a new task on an existing container.
//
// Any leftover task from a previous run is cleaned up first. The container
// must already exist; use [Container.create] for initial creation.
func (c *Container) Start(ctx context.Context) error {
	ctr, err := c.client.LoadContainer(ctx, c.id)
	if err != nil {
		return err
	}

	// Delete any stale task left over from a prior run.
	if task, err := ctr.Task(ctx, nil); err == nil {
		task.Kill(ctx, syscall.SIGKILL)
		task.Delete(ctx, containerd.WithProcessKill)
	}

	return c.startTask(ctx, ctr)
}

// Creates the containerd container with the Crucible security baseline.
//
// Spec options are applied in a strict layering order that is a security
// invariant — do not reorder:
//
//  1. [constrictionOpts] — absolute-zero baseline. The entire OCI spec is
//     constructed from scratch with maximum isolation and zero resources.
//     No containerd defaults are used.
//  2. [withImageArgs] — extracts only entrypoint, cmd, and working directory
//     from the image config. User, env, and supplementary GIDs from the
//     image are deliberately ignored.
//  3. [oci.WithHostname] — sets the UTS hostname.
//  4. extraOpts — affordance-derived options that grant every resource and
//     capability the workload requires: syscalls, memory, CPU, PIDs, mounts,
//     environment variables, rlimits, and so on.
//
// A container with no affordances (empty extraOpts) is maximally isolated
// but cannot run any meaningful workload: seccomp blocks all syscalls
// except exit_group, no mounts exist, and no resources are allocated.
func (c *Container) create(ctx context.Context, hostname string, image containerd.Image, extraOpts ...oci.SpecOpts) (containerd.Container, error) {
	specOpts := constrictionOpts(c.hugepageSizes)
	specOpts = append(specOpts,
		withImageArgs(image),
		oci.WithHostname(hostname),
	)
	specOpts = append(specOpts, extraOpts...)

	return c.client.NewContainer(ctx, c.id,
		containerd.WithImage(image),
		containerd.WithSnapshotter(snapshotter),
		containerd.WithNewSnapshot(c.id, image),
		containerd.WithRuntime(ociRuntime, nil),
		containerd.WithNewSpec(specOpts...),
	)
}

// Creates a container from an empty filesystem (scratch).
//
// An empty snapshot is prepared directly in the snapshotter with no parent
// layers. The container has no image association. Spec options follow the
// same layering as [Container.create] except [withImageArgs] is skipped
// since there is no image config to extract.
func (c *Container) createScratch(ctx context.Context, hostname string, extraOpts ...oci.SpecOpts) (containerd.Container, error) {
	sn := c.client.SnapshotService(snapshotter)
	if _, err := sn.Prepare(ctx, c.id, ""); err != nil {
		return nil, err
	}

	specOpts := constrictionOpts(c.hugepageSizes)
	specOpts = append(specOpts, oci.WithHostname(hostname))
	specOpts = append(specOpts, extraOpts...)

	return c.client.NewContainer(ctx, c.id,
		containerd.WithSnapshotter(snapshotter),
		containerd.WithSnapshot(c.id),
		containerd.WithRuntime(ociRuntime, nil),
		containerd.WithNewSpec(specOpts...),
	)
}

// Starts the container's long-running task with no attached IO.
func (c *Container) startTask(ctx context.Context, ctr containerd.Container) error {
	task, err := ctr.NewTask(ctx, cio.NullIO)
	if err != nil {
		return err
	}
	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		return err
	}
	return nil
}

// Extracts only entrypoint, cmd, and working directory from the image config.
//
// Unlike containerd's [oci.WithImageConfig], this deliberately ignores user,
// environment variables, and supplementary GIDs from the image. The baseline
// spec controls those fields; only the command to run and its working
// directory come from the image.
func withImageArgs(image containerd.Image) oci.SpecOpts {
	return func(ctx context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		desc, err := image.Config(ctx)
		if err != nil {
			return err
		}
		raw, err := content.ReadBlob(ctx, image.ContentStore(), desc)
		if err != nil {
			return err
		}
		var img ociv1.Image
		if err := json.Unmarshal(raw, &img); err != nil {
			return err
		}
		if s.Process == nil {
			s.Process = &specs.Process{}
		}
		s.Process.Args = append(img.Config.Entrypoint, img.Config.Cmd...)
		if img.Config.WorkingDir != "" {
			s.Process.Cwd = img.Config.WorkingDir
		}
		return nil
	}
}

// Removes an existing container with this ID, if one exists.
//
// Any running task is killed and the container is deleted along with its
// snapshot. This is a no-op when no container with the ID is found.
func (c *Container) remove(ctx context.Context) {
	existing, err := c.client.LoadContainer(ctx, c.id)
	if err != nil {
		return
	}
	if task, err := existing.Task(ctx, nil); err == nil {
		task.Kill(ctx, syscall.SIGKILL)
		task.Delete(ctx, containerd.WithProcessKill)
	}
	existing.Delete(ctx, containerd.WithSnapshotCleanup)
}
