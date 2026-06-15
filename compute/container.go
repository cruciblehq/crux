package compute

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/mount"
	"github.com/containerd/containerd/v2/pkg/archive"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/crypto"
)

// Container runtime limits and defaults.
const (

	// Shell used by default.
	defaultShell = "/bin/sh"

	// Maximum UID or GID accepted by runc. Runc requires values to fit in
	// int32 even though the OCI spec field type is uint32.
	runcMaxID = 1<<31 - 1
)

// Suffixes appended to step IDs when naming container records and processes.
const (

	// Suffix appended to step IDs when naming container records.
	containerIDSuffix = "container"

	// Suffix appended to step IDs when naming exec processes.
	execIDSuffix = "exec"
)

// Grace period for SIGTERM before SIGKILL is sent during Stop.
const stopGraceTimeout = 10 * time.Second

// Overlay filesystem layout.
const (

	// Path of the rootfs directory within a containerd snapshot mount.
	rootfsPath = "rootfs"

	// Mount type for overlayfs.
	overlayMountType = "overlay"

	// Overlayfs mount option prefix for the upper (writable) layer directory.
	overlayUpperdirOpt = "upperdir="
)

// An isolation boundary with a persistent writable snapshot.
//
// A [Container] is created with [Client.Load] and holds a writable overlayfs
// snapshot over the base image. [Container.Start] starts the container's main
// process, which can be stopped with [Container.Stop]; the snapshot persists
// across multiple Start/Stop cycles. [Container.Exec] runs a command in the
// container, and [Container.Copy] applies a tar archive into the snapshot;
// both accumulate changes that persist in the snapshot. Once the desired state
// is reached, [Container.Commit] captures the snapshot as a new image. The
// container and snapshot are released by calling [Container.Destroy] when they
// are no longer needed.
type Container struct {
	namespace   string                       // Containerd namespace for all operations.
	conn        *containerd.Client           // Containerd connection borrowed from the creating Client.
	img         containerd.Image             // Current base image.
	config      *ocispec.ImageConfig         // OCI config override applied at Commit time; nil means use img's config.
	snapshotID  string                       // Key of the active writable snapshot in the store.
	snapshotter string                       // Snapshotter name ("overlayfs").
	oci         specs.Spec                   // Compiled security spec; used as the runtime spec base.
	mu          sync.Mutex                   // Serialises Start, Stop, Exec, and Copy calls.
	ctr         containerd.Container         // Running container record; nil when stopped.
	task        containerd.Task              // Running task (PID 1); nil when stopped.
	exitCh      <-chan containerd.ExitStatus // Exit status channel from task.Wait; nil when stopped.
}

// Starts the container's main process.
//
// The process command is derived from the image's Entrypoint and Cmd fields,
// which can be overridden via [Container.Configure] before calling Start. If
// the image has neither, the OCI runtime rejects the spec at start time. The
// container must not already be running; a second call returns an error. Use
// [Container.Stop] to terminate the process and [Container.Destroy] to release
// all resources.
func (c *Container) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.task != nil {
		return crex.Wrapf(ErrContainer, "container already started")
	}

	containerID := fmt.Sprintf("%s-%s", crypto.RandHex(idLen), containerIDSuffix)

	ctn, err := c.conn.NewContainer(ctx, containerID,
		containerd.WithImage(c.img),
		containerd.WithSnapshotter(c.snapshotter),
		containerd.WithSnapshot(c.snapshotID),
		containerd.WithNewSpec(
			applyOCISpec(c.oci),
			oci.WithImageConfig(c.img),
			withComputeFields(c.namespace, containerID),
		),
	)
	if err != nil {
		return crex.Wrapf(ErrContainer, "create container: %w", err)
	}

	task, err := ctn.NewTask(ctx, cio.NullIO)
	if err != nil {
		ctn.Delete(ctx)
		return crex.Wrapf(ErrContainer, "create task: %w", err)
	}

	exitCh, err := task.Wait(ctx)
	if err != nil {
		task.Delete(ctx)
		ctn.Delete(ctx)
		return crex.Wrapf(ErrContainer, "wait task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		ctn.Delete(ctx)
		return crex.Wrapf(ErrContainer, "start task: %w", err)
	}

	c.ctr = ctn
	c.task = task
	c.exitCh = exitCh
	return nil
}

// Stops the container's main process.
//
// Sends SIGTERM and waits up to [stopGraceTimeout] seconds for a graceful exit
// before sending SIGKILL. The writable snapshot is not affected; changes from
// [Container.Copy] and [Container.Exec] calls are preserved. Safe to call on
// an already-stopped container.
func (c *Container) Stop(ctx context.Context) error {
	c.mu.Lock()
	task, ctr, exitCh := c.task, c.ctr, c.exitCh
	c.task, c.ctr, c.exitCh = nil, nil, nil
	c.mu.Unlock()

	if task == nil {
		return nil
	}

	cleanupCtx := namespaces.WithNamespace(context.Background(), c.namespace)
	_ = task.Kill(cleanupCtx, syscall.SIGTERM)

	stopCtx, cancel := context.WithTimeout(cleanupCtx, stopGraceTimeout)
	defer cancel()
	select {
	case <-exitCh:
	case <-stopCtx.Done():
		_ = task.Kill(cleanupCtx, syscall.SIGKILL)
		<-exitCh
	}

	task.Delete(cleanupCtx)
	ctr.Delete(cleanupCtx)
	return nil
}

// Runs a command in the container.
//
// Execs the command into the container's running main process. The command is
// invoked via opts.Shell (or the default shell) as a single argument. Changes
// to the filesystem accumulate in the writable snapshot and persist for
// subsequent [Container.Exec] and [Container.Copy] calls. A non-zero exit code
// is returned as an error.
func (c *Container) Exec(ctx context.Context, command string, opts RuntimeOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.task == nil {
		return crex.Wrapf(ErrContainer, "container not started")
	}

	ctx = namespaces.WithNamespace(ctx, c.namespace)

	if err := c.runExec(ctx, opts.Stdout, opts.Stderr, command, &opts); err != nil {
		return err
	}
	return nil
}

// Applies an uncompressed tar archive to the container's writable snapshot.
//
// r must be an uncompressed tar archive. Entries are applied as absolute paths
// within the container filesystem, preserving all permissions, ownership, and
// timestamps. OCI whiteout entries follow standard layer semantics. Changes
// accumulate in the snapshot. The container's main process is not affected.
func (c *Container) Copy(ctx context.Context, r io.Reader) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	mounts, err := c.conn.SnapshotService(c.snapshotter).Mounts(ctx, c.snapshotID)
	if err != nil {
		return crex.Wrapf(ErrContainer, "get snapshot mounts: %w", err)
	}

	if err := mount.WithTempMount(ctx, mounts, func(root string) error {
		_, err := archive.Apply(ctx, root, r)
		return err
	}); err != nil {
		return crex.Wrapf(ErrContainer, "apply tar: %w", err)
	}

	return nil
}

// Stores cfg as the pending configuration to be applied at [Container.Commit].
//
// Does not create an image or affect the writable snapshot. Typically called
// after [Container.Inspect] to modify specific fields while preserving the
// rest. The change takes effect when [Container.Commit] is called.
func (c *Container) Configure(cfg ocispec.ImageConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = &cfg
}

// Creates a new image from the container's current state.
//
// If the snapshot has no filesystem changes and no pending configuration,
// returns the current image reference unchanged. If only a configuration
// change is pending, writes a config-only image with an empty-layer history
// entry. Otherwise commits the snapshot diff, incorporating any pending
// configuration. The reference can be passed to [Client.Load], [Client.Export],
// [Client.Extract], or [Client.Remove]. The live container is not affected.
// Records a history entry using by as the creator and the current time.
func (c *Container) Commit(ctx context.Context, by string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	mounts, err := c.conn.SnapshotService(c.snapshotter).Mounts(ctx, c.snapshotID)
	if err != nil {
		return "", crex.Wrapf(ErrContainer, "get snapshot mounts: %w", err)
	}
	if !overlayHasChanges(mounts) {
		if c.config == nil {
			return c.img.Name(), nil
		}
		return configureImage(ctx, c.conn, c.img, *c.config, by)
	}
	return commitSnapshotDiff(ctx, c.conn, c.img, c.snapshotID, c.snapshotter, c.config, by)
}

// Releases the container's writable snapshot.
//
// After Destroy returns, no other methods may be called on the container.
func (c *Container) Destroy(ctx context.Context) error {
	c.conn.SnapshotService(c.snapshotter).Remove(ctx, c.snapshotID)
	return nil
}

// Returns the container's current OCI image config.
//
// If a pending configuration has been set via [Container.Configure], that is
// returned directly. If not, the config is read from the base image in the
// content store. The caller owns the returned struct and may modify it before
// passing it back to [Container.Configure].
func (c *Container) Inspect(ctx context.Context) (ocispec.ImageConfig, error) {
	c.mu.Lock()
	pending := c.config
	img := c.img
	c.mu.Unlock()
	if pending != nil {
		return *pending, nil
	}
	_, cfg, err := readImageConfig(ctx, c.conn.ContentStore(), img.Target())
	if err != nil {
		return ocispec.ImageConfig{}, err
	}
	return cfg.Config, nil
}

// Whether an overlayfs upper directory contains any entries.
//
// Parses the upperdir option from mounts and reads the directory. On error,
// returns true to conservatively treat the snapshot as changed. Returns false
// if no overlay mount is present in the slice.
func overlayHasChanges(mounts []mount.Mount) bool {
	for _, m := range mounts {
		if m.Type != overlayMountType {
			continue
		}
		for _, opt := range m.Options {
			if !strings.HasPrefix(opt, overlayUpperdirOpt) {
				continue
			}
			entries, err := os.ReadDir(strings.TrimPrefix(opt, overlayUpperdirOpt))
			if err != nil {
				return true
			}
			return len(entries) > 0
		}
	}
	return false
}

// Execs command into the running task, routing output through an [execIO].
func (c *Container) runExec(ctx context.Context, stdout, stderr io.Writer, command string, opts *RuntimeOptions) error {
	execID := fmt.Sprintf("%s-%s", crypto.RandHex(idLen), execIDSuffix)
	eio := newExecIO(stdout, stderr)
	defer eio.close()

	creator, err := eio.creator()
	if err != nil {
		return crex.Wrapf(ErrContainer, "create log IO: %w", err)
	}

	spec, err := c.ctr.Spec(ctx)
	if err != nil {
		return crex.Wrapf(ErrContainer, "get spec: %w", err)
	}

	process, err := c.task.Exec(ctx, execID, buildExecProcess(spec, command, opts), creator)
	if err != nil {
		return crex.Wrapf(ErrContainer, "exec: %w", err)
	}

	exitCh, err := process.Wait(ctx)
	if err != nil {
		process.Delete(ctx)
		return crex.Wrapf(ErrContainer, "exec wait: %w", err)
	}

	if err := process.Start(ctx); err != nil {
		process.Delete(ctx)
		return crex.Wrapf(ErrContainer, "exec start: %w", err)
	}

	var execErr error
	select {
	case status := <-exitCh:
		if err := status.Error(); err != nil {
			execErr = crex.Wrapf(ErrContainer, "exec: %w", err)
		} else if code := status.ExitCode(); code != 0 {
			execErr = crex.Wrapf(ErrContainer, "command exited with code %d", code)
		}
	case <-ctx.Done():
		process.Kill(ctx, syscall.SIGTERM)
		execErr = ctx.Err()
	}

	eio.flush(process.IO())
	process.Delete(ctx)

	return execErr
}

// Builds a process spec for exec from the container's base process.
//
// Overrides args, env, workdir, and user from command and opts.
func buildExecProcess(base *specs.Spec, command string, opts *RuntimeOptions) *specs.Process {
	proc := *base.Process
	proc.Terminal = false

	shell := defaultShell
	if opts.Shell != "" {
		shell = opts.Shell
	}
	proc.Args = []string{shell, "-c", command}

	if len(opts.Env) > 0 {
		env := make([]string, 0, len(opts.Env))
		for k, v := range opts.Env {
			env = append(env, k+"="+v)
		}
		proc.Env = MergeEnv(proc.Env, env)
	}

	if opts.Workdir != "" {
		proc.Cwd = opts.Workdir
	}

	if opts.User != "" {
		uid, gid, err := parseUserSpec(opts.User)
		if err == nil {
			proc.User.UID = uid
			if gid != nil {
				proc.User.GID = *gid
			}
		}
	}

	return &proc
}

// Parses a user spec in "uid" or "uid:gid" format.
//
// The uid is required. The gid is optional; a nil gid indicates no colon was
// present. Both components must satisfy the [parseID] constraint.
func parseUserSpec(user string) (uid uint32, gid *uint32, err error) {
	parts := strings.SplitN(user, ":", 2)
	uid, err = parseID(parts[0])
	if err != nil {
		return 0, nil, crex.Wrapf(ErrContainer, "invalid uid %q: %w", user, err)
	}

	if len(parts) == 2 {
		g, err := parseID(parts[1])
		if err != nil {
			return 0, nil, crex.Wrapf(ErrContainer, "invalid gid %q: %w", user, err)
		}
		gid = &g
	}
	return uid, gid, nil
}

// Parses a decimal integer as a process ID.
//
// Values above MaxInt32 are rejected to match runc's requirement, even though
// they are valid uint32 values.
func parseID(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	if v > runcMaxID {
		return 0, errors.New("exceeds runc maximum")
	}
	return uint32(v), nil
}

// Copies src into the blank spec produced by [containerd.WithNewSpec] so the
// compiled security spec becomes the starting point for the full container spec.
//
// Top-level pointer fields (Root, Process, Linux) are shallow-copied into new
// structs so that subsequent SpecOpts can modify them without mutating the
// original security spec.
func applyOCISpec(src specs.Spec) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		*s = src
		if src.Root != nil {
			r := *src.Root
			s.Root = &r
		}
		if src.Process != nil {
			p := *src.Process
			s.Process = &p
		}
		if src.Linux != nil {
			l := *src.Linux
			s.Linux = &l
		}
		return nil
	}
}

// Sets the two fields the affordance package leaves empty because only the
// container runtime knows them: the rootfs path (always "rootfs" under
// containerd's overlayfs snapshotter) and the cgroup path.
func withComputeFields(ns, id string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		s.Root.Path = rootfsPath
		s.Linux.CgroupsPath = filepath.Join("/", ns, id)
		return nil
	}
}
