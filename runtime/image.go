package runtime

import (
	"context"
	"fmt"
	"os"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/containerd/errdefs"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/reference"
)

// An OCI image within the container runtime's image store.
//
// The containerd namespace is derived from the registry component of the
// resource reference, providing isolation between registries. The image is
// tagged as "namespace/name:version".
type Image struct {
	registry  string // Containerd namespace (from registry).
	namespace string // Resource namespace.
	name      string // Resource name.
	version   string // Image version.
}

// Returns the image reference as "namespace/name".
func (img *Image) ref() string {
	return fmt.Sprintf("%s/%s", img.namespace, img.name)
}

// Returns the image tag as "namespace/name:version".
func (img *Image) tag() string {
	return fmt.Sprintf("%s:%s", img.ref(), img.version)
}

// Creates an [Image] from a parsed resource identifier and version.
//
// The registry component of the identifier is used as the containerd
// namespace. The image name is derived from the identifier's namespace
// and name.
func NewImage(id *reference.Identifier, version string) *Image {
	return &Image{
		registry:  id.Registry(),
		namespace: id.Namespace(),
		name:      id.Name(),
		version:   version,
	}
}

// Imports an OCI image tarball into the container runtime's image store.
//
// The tarball at path is loaded into containerd and tagged within the
// image's namespace.
func (img *Image) Import(ctx context.Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return crex.Wrap(ErrImageFileOpen, err)
	}
	defer f.Close()

	c, err := newContainerdClient(img.registry)
	if err != nil {
		return crex.Wrap(ErrImageImport, err)
	}
	defer c.Close()

	_, err = c.Import(ctx, f, containerd.WithImageRefTranslator(func(_ string) string {
		return img.tag()
	}))
	if err != nil {
		return crex.Wrap(ErrImageImport, err)
	}

	return nil
}

// Destroys the image and all its containers.
//
// Containers created from this image are discovered by querying containerd
// and destroyed first. The image is then removed from the image store.
func (img *Image) Destroy(ctx context.Context) error {
	c, err := newContainerdClient(img.registry)
	if err != nil {
		return crex.Wrap(ErrImageDestroy, err)
	}
	defer c.Close()

	filter := fmt.Sprintf("image==%s", img.tag())
	ctrs, err := c.Containers(ctx, filter)
	if err != nil {
		return crex.Wrap(ErrImageDestroy, err)
	}

	for _, ctr := range ctrs {
		if task, taskErr := ctr.Task(ctx, nil); taskErr == nil {
			task.Kill(ctx, syscall.SIGKILL)
			task.Delete(ctx, containerd.WithProcessKill)
		}
		if err := ctr.Delete(ctx, containerd.WithSnapshotCleanup); err != nil && !errdefs.IsNotFound(err) {
			return crex.Wrap(ErrImageDestroy, err)
		}
	}

	if err := c.ImageService().Delete(ctx, img.tag()); err != nil && !errdefs.IsNotFound(err) {
		return crex.Wrap(ErrImageDestroy, err)
	}

	return nil
}

// Creates and starts a container from this image.
//
// If id is empty, the image name is used as the default container
// identifier. Any existing container with the same identifier is
// cleaned up first. The container runs detached.
func (img *Image) Start(ctx context.Context, id string) (*Container, error) {
	if id == "" {
		id = img.name
	}

	client, err := newContainerdClient(img.registry)
	if err != nil {
		return nil, crex.Wrap(ErrContainerStart, err)
	}
	defer client.Close()

	cleanupStaleContainer(ctx, client, id)

	image, err := client.GetImage(ctx, img.tag())
	if err != nil {
		return nil, crex.Wrap(ErrContainerStart, err)
	}

	ctr, err := client.NewContainer(ctx, id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)
	if err != nil {
		return nil, crex.Wrap(ErrContainerStart, err)
	}

	if err := startTask(ctx, ctr); err != nil {
		ctr.Delete(ctx, containerd.WithSnapshotCleanup)
		return nil, crex.Wrap(ErrContainerStart, err)
	}

	return NewContainer(img.registry, id), nil
}

// Stops the container, re-imports the image from a new tarball, and
// restarts the container with the same identifier.
func (img *Image) Update(ctx context.Context, c *Container, path string) error {
	if err := c.Stop(ctx); err != nil {
		return err
	}
	if err := img.Import(ctx, path); err != nil {
		return err
	}
	_, err := img.Start(ctx, c.id)
	return err
}

// Removes a leftover container and its task from a previous run.
func cleanupStaleContainer(ctx context.Context, client *containerd.Client, id string) {
	existing, err := client.LoadContainer(ctx, id)
	if err != nil {
		return
	}
	if task, err := existing.Task(ctx, nil); err == nil {
		task.Kill(ctx, syscall.SIGKILL)
		task.Delete(ctx, containerd.WithProcessKill)
	}
	existing.Delete(ctx, containerd.WithSnapshotCleanup)
}

// Creates and starts a task for the container in detached mode.
func startTask(ctx context.Context, ctr containerd.Container) error {
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
