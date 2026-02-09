package runtime

import (
	"context"
	"fmt"
	"os"
	"sync"

	containerd "github.com/containerd/containerd/v2/client"
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
	registry   string       // Containerd namespace (from registry).
	namespace  string       // Resource namespace.
	name       string       // Resource name.
	version    string       // Image version.
	mu         sync.Mutex   // Guards concurrent access to containers.
	containers []*Container // Containers started from this image.
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
// Every container started from this image is destroyed first. The image
// is then removed from containerd's image store. After destruction no new
// containers can be started from this image.
func (img *Image) Destroy(ctx context.Context) error {
	img.mu.Lock()
	remaining := make([]*Container, len(img.containers))
	copy(remaining, img.containers)
	img.mu.Unlock()

	for _, c := range remaining {
		if err := c.Destroy(ctx); err != nil {
			return crex.Wrap(ErrImageDestroy, err)
		}
	}

	c, err := newContainerdClient(img.registry)
	if err != nil {
		return crex.Wrap(ErrImageDestroy, err)
	}
	defer c.Close()

	if err := c.ImageService().Delete(ctx, img.tag()); err != nil && !errdefs.IsNotFound(err) {
		return crex.Wrap(ErrImageDestroy, err)
	}

	return nil
}

// Creates and starts a container from this image.
//
// If id is empty, the image name is used as the default container
// identifier. The container runs detached.
func (img *Image) Start(ctx context.Context, id string) (*Container, error) {
	if id == "" {
		id = img.name
	}

	c := &Container{image: img, id: id}
	if err := c.Start(ctx); err != nil {
		return nil, err
	}

	img.mu.Lock()
	img.containers = append(img.containers, c)
	img.mu.Unlock()

	return c, nil
}
