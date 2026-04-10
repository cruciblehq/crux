package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	goruntime "runtime"
	"strings"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/transfer/archive"
	timage "github.com/containerd/containerd/v2/core/transfer/image"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/containerd/errdefs"
	"github.com/containerd/platforms"
	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

const (

	// Snapshotter used for container filesystems. containerd runs as root
	// inside the VM, so the native overlayfs kernel module is available.
	snapshotter = "overlayfs"

	// OCI runtime shim for running containers.
	ociRuntime = "io.containerd.runc.v2"
)

// Manages the containerd client and provides image and container operations.
type Runtime struct {
	client *containerd.Client // Containerd client for managing containers and images.
}

// Creates a runtime connected to the containerd socket at the given address.
//
// The namespace scopes all containerd operations to a single tenant. The
// runtime must be closed when no longer needed.
func New(address, namespace string) (*Runtime, error) {
	client, err := containerd.New(address, containerd.WithDefaultNamespace(namespace))
	if err != nil {
		return nil, crex.Wrap(ErrRuntime, err)
	}
	return &Runtime{client: client}, nil
}

// Closes the containerd client connection.
func (rt *Runtime) Close() error {
	return rt.client.Close()
}

// Imports an OCI archive, unpacks it for the target platform, and starts
// a container.
//
// The archive is transferred server-side into containerd's content store,
// tagged with a deterministic name derived from the path, and the layers
// for the target platform are unpacked into the snapshotter. A container
// is created with a fresh snapshot and a long-running task (sleep infinity)
// is started so that subsequent Exec calls have a running process to attach
// to. Any existing container with the same ID is removed before the new one
// is created. Building for a platform other than the host requires
// QEMU / binfmt_misc support in the kernel.
func (rt *Runtime) StartContainer(ctx context.Context, path string, id string, platform string) (*Container, error) {
	tag := imageTag(path)

	if err := rt.transferImage(ctx, path, tag, platform); err != nil {
		return nil, crex.Wrap(ErrRuntime, err)
	}

	c := &Container{
		client:   rt.client,
		id:       id,
		platform: platform,
	}

	// Remove any stale container from a previous build with the same ID.
	c.remove(ctx)

	image, err := rt.resolveImage(ctx, tag, platform)
	if err != nil {
		return nil, crex.Wrap(ErrRuntime, err)
	}

	ctr, err := c.create(ctx, id, image,
		oci.WithProcessArgs("sleep", "infinity"),
	)
	if err != nil {
		return nil, crex.Wrap(ErrRuntime, err)
	}

	if err := c.startTask(ctx, ctr); err != nil {
		ctr.Delete(ctx, containerd.WithSnapshotCleanup)
		return nil, crex.Wrap(ErrRuntime, err)
	}

	return c, nil
}

// Creates a container from an empty filesystem (scratch).
//
// No image is imported or unpacked. The container starts with a completely
// empty rootfs. Any existing container with the same ID is removed before
// the new one is created.
func (rt *Runtime) StartScratchContainer(ctx context.Context, id string, platform string) (*Container, error) {
	c := &Container{
		client:   rt.client,
		id:       id,
		platform: platform,
	}

	c.remove(ctx)

	ctr, err := c.createScratch(ctx, id,
		oci.WithProcessArgs("sleep", "infinity"),
	)
	if err != nil {
		return nil, crex.Wrap(ErrRuntime, err)
	}

	if err := c.startTask(ctx, ctr); err != nil {
		ctr.Delete(ctx, containerd.WithSnapshotCleanup)
		return nil, crex.Wrap(ErrRuntime, err)
	}

	return c, nil
}

// Creates a scratch container without starting a task.
//
// Used for the export-only path when the recipe has no stages. The
// container gets an empty snapshot but no running process, since a scratch
// rootfs has no binaries to execute. The returned container can be exported
// directly without calling Stop first.
func (rt *Runtime) CreateScratchContainer(ctx context.Context, id string, platform string) (*Container, error) {
	c := &Container{
		client:   rt.client,
		id:       id,
		platform: platform,
	}

	c.remove(ctx)

	if _, err := c.createScratch(ctx, id); err != nil {
		return nil, crex.Wrap(ErrRuntime, err)
	}

	return c, nil
}

// Transfers an OCI archive into containerd's content store server-side.
//
// The archive is streamed to containerd which imports it, stores it under
// the given tag, and unpacks the layers for the target platform into the
// snapshotter. The entire operation runs inside the containerd process,
// so crux does not need mount privileges.
func (rt *Runtime) transferImage(ctx context.Context, path, tag, platform string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	p, err := platforms.Parse(platform)
	if err != nil {
		return err
	}

	src := archive.NewImageImportStream(fh, "")
	dest := timage.NewStore(tag, timage.WithUnpack(p, snapshotter))

	return rt.client.Transfer(ctx, src, dest)
}

// Looks up a tagged image and selects the manifest for the given platform.
//
// Multi-platform images contain manifests for multiple architectures. This
// method selects one, so that subsequent operations target the correct
// architecture.
func (rt *Runtime) resolveImage(ctx context.Context, tag, platform string) (containerd.Image, error) {
	p, err := platforms.Parse(platform)
	if err != nil {
		return nil, err
	}

	img, err := rt.client.ImageService().Get(ctx, tag)
	if err != nil {
		return nil, err
	}

	return containerd.NewImageWithPlatform(rt.client, img, platforms.Only(p)), nil
}

// Produces a containerd image tag from an archive path.
//
// The path is hashed to produce a tag that is always valid for OCI references
// regardless of which characters the path contains.
func imageTag(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("import/%s:latest", hex.EncodeToString(h[:]))
}

// Returns the default OCI platform for the host architecture.
func defaultPlatform() string {
	return "linux/" + goruntime.GOARCH
}

// ImageTag produces a containerd image tag for a resource reference and version.
//
// The tag is formatted as "ref:version", which is used to import and retrieve
// images by resource identity.
func ImageTag(ref, version string) string {
	return ref + ":" + version
}

// ContainerID produces a container identifier from a resource reference.
//
// Slashes in the reference are replaced with dashes to produce a valid
// containerd container ID.
func ContainerID(ref string) string {
	return strings.ReplaceAll(ref, "/", "-")
}

// Imports an OCI archive, tags it under the given name, and unpacks it for
// the host platform.
//
// The archive is transferred server-side into containerd's content store,
// tagged with the provided name, and the layers are unpacked into the
// snapshotter.
func (rt *Runtime) ImportImage(ctx context.Context, path, tag string) error {
	platform := defaultPlatform()
	if err := rt.transferImage(ctx, path, tag, platform); err != nil {
		return crex.Wrap(ErrRuntime, err)
	}
	return nil
}

// Starts a container from a previously imported image tag.
//
// The operation is idempotent: if the container is already running it is
// left untouched; if the container exists but has no active task a new
// task is started on the existing snapshot; otherwise a new container is
// created from the image. Affordances describe the capabilities the service
// requires; they will be resolved into OCI spec options that configure the
// container's sandbox.
func (rt *Runtime) StartFromTag(ctx context.Context, tag, id string, affordances []manifest.Ref) (*Container, error) {
	platform := defaultPlatform()

	c := &Container{
		client:   rt.client,
		id:       id,
		platform: platform,
	}

	status, err := c.Status(ctx)
	if err != nil {
		return nil, crex.Wrap(ErrRuntime, err)
	}

	switch status {
	case ContainerRunning:
		return c, nil

	case ContainerStopped:
		if err := c.Start(ctx); err != nil {
			return nil, crex.Wrap(ErrRuntime, err)
		}
		return c, nil

	default:
		image, err := rt.resolveImage(ctx, tag, platform)
		if err != nil {
			return nil, crex.Wrap(ErrRuntime, err)
		}

		ctr, err := c.create(ctx, id, image)
		if err != nil {
			return nil, crex.Wrap(ErrRuntime, err)
		}

		if err := c.startTask(ctx, ctr); err != nil {
			ctr.Delete(ctx, containerd.WithSnapshotCleanup)
			return nil, crex.Wrap(ErrRuntime, err)
		}

		return c, nil
	}
}

// Removes an image and all containers created from it.
//
// Containers are discovered by querying containerd for records whose image
// field matches the tag. Each container's task is killed before the container
// and its snapshot are deleted.
func (rt *Runtime) DestroyImage(ctx context.Context, tag string) error {
	ctrs, err := rt.client.Containers(ctx, fmt.Sprintf("image==%s", tag))
	if err != nil {
		return crex.Wrap(ErrRuntime, err)
	}

	for _, ctr := range ctrs {
		if task, taskErr := ctr.Task(ctx, nil); taskErr == nil {
			task.Kill(ctx, syscall.SIGKILL)
			task.Delete(ctx, containerd.WithProcessKill)
		}
		if err := ctr.Delete(ctx, containerd.WithSnapshotCleanup); err != nil && !errdefs.IsNotFound(err) {
			return crex.Wrap(ErrRuntime, err)
		}
	}

	if err := rt.client.ImageService().Delete(ctx, tag); err != nil && !errdefs.IsNotFound(err) {
		return crex.Wrap(ErrRuntime, err)
	}

	return nil
}

// Returns a handle for an existing container.
//
// The container is not loaded or verified; the handle is a lightweight
// reference that resolves the container lazily on subsequent calls.
func (rt *Runtime) Container(id string) *Container {
	return &Container{
		client:   rt.client,
		id:       id,
		platform: defaultPlatform(),
	}
}
