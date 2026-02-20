package runtime

import (
	"context"
	"fmt"
	"os"
	goruntime "runtime"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/images/archive"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/containerd/errdefs"
	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/reference"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (

	// Snapshotter used for container filesystems.
	containerdSnapshotter = "overlayfs"

	// OCI runtime shim for running containers.
	containerdRuntime = "io.containerd.runc.v2"
)

// OCI platform for the build container.
//
// Currently hardcoded to the host architecture. Multi-architecture image
// support (building for both amd64 and arm64 via QEMU/Rosetta emulation
// and producing an OCI image index) is not yet implemented. All images
// are single-platform for the architecture of the machine running the build.
var containerPlatform = "linux/" + goruntime.GOARCH

// An OCI image within the container runtime's image store.
//
// The containerd namespace is derived from the registry component of the
// resource reference, providing isolation between registries. The image is
// tagged as "namespace/name:version".
type Image struct {
	client    *containerd.Client // Shared containerd gRPC connection.
	registry  string             // Containerd namespace (from registry).
	namespace string             // Resource namespace.
	name      string             // Resource name.
	version   string             // Image version.
}

// Returns the image reference as "namespace/name".
func (img *Image) ref() string {
	return fmt.Sprintf("%s/%s", img.namespace, img.name)
}

// Returns the image tag as "namespace/name:version".
func (img *Image) tag() string {
	return fmt.Sprintf("%s:%s", img.ref(), img.version)
}

// Returns a containerd filter expression that matches containers using this image.
func (img *Image) filter() string {
	return fmt.Sprintf("image==%s", img.tag())
}

// Creates an [Image] from a parsed resource identifier and version.
//
// The client is a shared containerd gRPC connection whose lifecycle is
// managed by the caller. The registry component of the identifier is
// used as the containerd namespace. The image name is derived from the
// identifier's namespace and name.
func NewImage(client *containerd.Client, id *reference.Identifier, version string) *Image {
	return &Image{
		client:    client,
		registry:  id.Hostname(),
		namespace: id.Namespace(),
		name:      id.Name(),
		version:   version,
	}
}

// Imports an OCI image tarball into the container runtime's image store.
//
// The archive may contain images under their original references (e.g.
// docker.io/library/alpine:3.21). After importing the content blobs,
// a new image record is created under [Image.tag] pointing at the same
// target descriptor. The original reference is removed so only our tag
// remains. The image is then unpacked into the snapshotter so it is ready
// for container creation.
func (img *Image) Import(ctx context.Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return crex.Wrap(ErrImageFileOpen, err)
	}
	defer f.Close()

	imported, err := img.client.Import(ctx, f)
	if err != nil {
		return crex.Wrap(ErrImageImport, err)
	}

	rec, err := validateImport(imported)
	if err != nil {
		return err
	}

	if err := img.retag(ctx, rec); err != nil {
		return err
	}

	return img.unpack(ctx)
}

// Validates that the import produced exactly one image record.
func validateImport(imported []images.Image) (images.Image, error) {
	if len(imported) == 0 {
		return images.Image{}, crex.Wrap(ErrImageImport, ErrImageEmpty)
	}
	if len(imported) > 1 {
		return images.Image{}, crex.Wrap(ErrImageImport, ErrImageMultiple)
	}
	return imported[0], nil
}

// Retags an imported image record under [Image.tag].
//
// The containerd image import process creates new image records under their
// original references (e.g. docker.io/library/alpine:3.21), but we need them
// under our own tag (e.g. my-registry/my-namespace/my-service:version), since
// that's what's used for container creation and lookup. This creates a new
// image record pointing at the same target descriptor, then removes the
// original reference.
func (img *Image) retag(ctx context.Context, rec images.Image) error {
	is := img.client.ImageService()
	tag := img.tag()

	if _, err := is.Create(ctx, images.Image{
		Name:   tag,
		Target: rec.Target,
	}); err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return crex.Wrap(ErrImageImport, err)
		}
		if _, err := is.Update(ctx, images.Image{
			Name:   tag,
			Target: rec.Target,
		}, "target"); err != nil {
			return crex.Wrap(ErrImageImport, err)
		}
	}

	// Remove the original reference if it differs from our tag.
	if rec.Name != tag {
		_ = is.Delete(ctx, rec.Name)
	}

	return nil
}

// Unpacks the image layers into the snapshotter so they are ready for
// container creation.
func (img *Image) unpack(ctx context.Context) error {
	tagged, err := img.client.GetImage(ctx, img.tag())
	if err != nil {
		return crex.Wrap(ErrImageImport, err)
	}
	if err := tagged.Unpack(ctx, containerdSnapshotter); err != nil {
		return crex.Wrap(ErrImageImport, err)
	}

	return nil
}

// Destroys the image and all its containers.
//
// Containers created from this image are discovered by querying containerd
// and destroyed first. The image is then removed from the image store.
func (img *Image) Destroy(ctx context.Context) error {
	ctrs, err := img.client.Containers(ctx, img.filter())
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

	if err := img.client.ImageService().Delete(ctx, img.tag()); err != nil && !errdefs.IsNotFound(err) {
		return crex.Wrap(ErrImageDestroy, err)
	}

	return nil
}

// Creates and starts a container from this image.
//
// If id is empty, the image name is used as the default container identifier.
// Any existing container with the same identifier is cleaned up first. The
// container runs detached.
func (img *Image) Start(ctx context.Context, id string) (*Container, error) {
	if id == "" {
		id = img.name
	}

	cleanupStaleContainer(ctx, img.client, id)

	image, err := img.client.GetImage(ctx, img.tag())
	if err != nil {
		return nil, crex.Wrap(ErrContainerStart, err)
	}

	ctr, err := createContainer(ctx, img.client, id, image)
	if err != nil {
		return nil, crex.Wrap(ErrContainerStart, err)
	}

	if err := startTask(ctx, ctr); err != nil {
		ctr.Delete(ctx, containerd.WithSnapshotCleanup)
		return nil, crex.Wrap(ErrContainerStart, err)
	}

	return NewContainer(img.client, img.registry, id), nil
}

// Stops the container, re-imports the image from a new tarball, and restarts
// the container with the same identifier.
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

// Sets the entrypoint on the image's OCI config.
//
// The image manifest and config are read from the content store, the
// entrypoint is updated, and the modified config and manifest are written
// back. The image record is then pointed at the new manifest. The
// container must be committed before calling this method. If entrypoint
// is nil or empty, no changes are made.
func (img *Image) SetEntrypoint(ctx context.Context, entrypoint []string) error {
	if len(entrypoint) == 0 {
		return nil
	}

	return updateImageConfig(ctx, img.client, img.tag(), func(config *ocispec.Image) {
		config.Config.Entrypoint = entrypoint
		config.Config.Cmd = nil
	})
}

// Exports the image as an OCI tar archive.
//
// The archive is written to the specified path in OCI image layout format,
// containing all layers, the manifest, and the config. The resulting file
// is suitable for distribution or import into another runtime.
func (img *Image) Export(ctx context.Context, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return crex.Wrap(ErrImageExport, err)
	}
	defer f.Close()

	opt := archive.WithImage(img.client.ImageService(), img.tag())

	if err := img.client.Export(ctx, f, opt); err != nil {
		return crex.Wrap(ErrImageExport, err)
	}

	return nil
}

// Creates a container with a fresh snapshot and OCI spec from the image.
//
// The container shares the host network namespace for outbound access and runs
// "sleep infinity" as PID 1 to stay alive for exec commands.
func createContainer(ctx context.Context, client *containerd.Client, id string, image containerd.Image) (containerd.Container, error) {
	return client.NewContainer(ctx, id,
		containerd.WithImage(image),
		containerd.WithSnapshotter(containerdSnapshotter),
		containerd.WithNewSnapshot(id, image),
		containerd.WithRuntime(containerdRuntime, nil),
		containerd.WithNewSpec(
			oci.WithDefaultSpecForPlatform(containerPlatform),
			oci.WithImageConfig(image),
			oci.WithHostNamespace(specs.NetworkNamespace),
			oci.WithHostResolvconf,
			oci.WithProcessArgs("sleep", "infinity"),
		),
	)
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
