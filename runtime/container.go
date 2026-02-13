package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/pkg/rootfs"
	"github.com/containerd/errdefs"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// A container instance within the container runtime.
//
// Use [NewContainer] to construct a handle for an existing container, or
// [Image.Start] to create and start one.
type Container struct {
	client   *containerd.Client // Shared containerd gRPC connection.
	registry string             // Containerd namespace.
	id       string             // Container identifier.
}

// Creates a [Container] handle for an existing container.
//
// The client is a shared containerd gRPC connection whose lifecycle is
// managed by the caller. The registry is the containerd namespace (the
// registry host authority). The id is the container identifier within
// that namespace. This does not create or start anything in the runtime.
func NewContainer(client *containerd.Client, registry, id string) *Container {
	return &Container{client: client, registry: registry, id: id}
}

// Stops the container's task.
//
// The running task is killed and deleted. The container metadata is
// preserved. Stop is idempotent; calling it on an already-stopped
// container is not an error.
func (c *Container) Stop(ctx context.Context) error {
	ctr, err := c.client.LoadContainer(ctx, c.id)
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
	ctr, loadErr := c.client.LoadContainer(ctx, c.id)
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
	return containerExec(ctx, c.client, c.registry, c.id, opts, command, args...)
}

// Queries the current state of the container.
//
// Returns [StateRunning] if the task is running, [StateStopped] if the
// container exists but has no running task, or [StateNotCreated] if the
// container does not exist.
func (c *Container) Status(ctx context.Context) (State, error) {
	ctr, err := c.client.LoadContainer(ctx, c.id)
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

// Commits the container's filesystem changes to its image.
//
// The diff between the container's active snapshot and its parent is computed
// by the containerd diff service and stored in the content store as a new
// compressed layer blob. The image manifest and config are then updated to
// include the new layer, and the image record is pointed at the new manifest.
// The container should be stopped before committing to ensure all filesystem
// writes have completed.
func (c *Container) Commit(ctx context.Context) error {
	ctr, err := c.client.LoadContainer(ctx, c.id)
	if err != nil {
		return crex.Wrap(ErrContainerCommit, err)
	}

	info, err := ctr.Info(ctx)
	if err != nil {
		return crex.Wrap(ErrContainerCommit, err)
	}

	if err := commitSnapshot(ctx, c.client, info); err != nil {
		return crex.Wrap(ErrContainerCommit, err)
	}

	return nil
}

// Computes the snapshot diff, appends it as a new layer to the image
// manifest and config, and updates the image record.
func commitSnapshot(ctx context.Context, client *containerd.Client, info containers.Container) error {
	diffDesc, err := rootfs.CreateDiff(ctx,
		info.SnapshotKey,
		client.SnapshotService(info.Snapshotter),
		client.DiffService(),
	)
	if err != nil {
		return err
	}

	diffID, err := images.GetDiffID(ctx, client.ContentStore(), diffDesc)
	if err != nil {
		return err
	}

	return appendLayer(ctx, client, info.Image, diffDesc, diffID)
}

// Reads the image's manifest and config, appends the layer descriptor and diff
// ID, writes the updated blobs, and points the image record at the new manifest.
func appendLayer(ctx context.Context, client *containerd.Client, imageName string, layer ocispec.Descriptor, diffID digest.Digest) error {
	cs := client.ContentStore()
	is := client.ImageService()

	img, err := is.Get(ctx, imageName)
	if err != nil {
		return err
	}

	manifest, err := readManifest(ctx, cs, img.Target)
	if err != nil {
		return err
	}

	config, err := readConfig(ctx, cs, manifest.Config)
	if err != nil {
		return err
	}

	config.RootFS.DiffIDs = append(config.RootFS.DiffIDs, diffID)
	manifest.Layers = append(manifest.Layers, layer)

	newConfigDesc, err := writeJSON(ctx, cs, manifest.Config.MediaType, config, imageName+"-config")
	if err != nil {
		return err
	}
	manifest.Config = newConfigDesc

	manifestLabels := manifestGCLabels(manifest)
	newManifestDesc, err := writeJSON(ctx, cs, img.Target.MediaType, manifest, imageName+"-manifest", content.WithLabels(manifestLabels))
	if err != nil {
		return err
	}

	img.Target = newManifestDesc
	_, err = is.Update(ctx, img, "target")
	return err
}

// Reads and unmarshals an OCI manifest from the content store.
func readManifest(ctx context.Context, cs content.Store, desc ocispec.Descriptor) (ocispec.Manifest, error) {
	b, err := content.ReadBlob(ctx, cs, desc)
	if err != nil {
		return ocispec.Manifest{}, err
	}
	var m ocispec.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return ocispec.Manifest{}, err
	}
	return m, nil
}

// Reads and unmarshals an OCI image config from the content store.
func readConfig(ctx context.Context, cs content.Store, desc ocispec.Descriptor) (ocispec.Image, error) {
	b, err := content.ReadBlob(ctx, cs, desc)
	if err != nil {
		return ocispec.Image{}, err
	}
	var img ocispec.Image
	if err := json.Unmarshal(b, &img); err != nil {
		return ocispec.Image{}, err
	}
	return img, nil
}

// Marshals a value to JSON and writes it to the content store, returning
// a descriptor for the written blob.
func writeJSON(ctx context.Context, cs content.Store, mediaType string, v any, ref string, opts ...content.Opt) (ocispec.Descriptor, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	desc := ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digest.FromBytes(b),
		Size:      int64(len(b)),
	}
	if err := content.WriteBlob(ctx, cs, ref, bytes.NewReader(b), desc, opts...); err != nil {
		return ocispec.Descriptor{}, err
	}
	return desc, nil
}

// Computes containerd GC reference labels for a manifest's children.
//
// These labels allow containerd's garbage collector to trace reachability from
// the manifest blob to its config and layer blobs. Without them, content
// written by [appendLayer] would be eligible for collection as soon as any
// concurrent GC pass runs. The label format mirrors what containerd's own
// import handler ([images.SetChildrenLabels]) produces.
func manifestGCLabels(m ocispec.Manifest) map[string]string {
	labels := map[string]string{
		"containerd.io/gc.ref.content.config": m.Config.Digest.String(),
	}
	for i, layer := range m.Layers {
		key := fmt.Sprintf("containerd.io/gc.ref.content.l.%d", i)
		labels[key] = layer.Digest.String()
	}
	return labels
}
