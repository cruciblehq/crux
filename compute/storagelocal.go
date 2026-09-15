package compute

import (
	"context"
	"fmt"
	"os"

	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/crypto"
)

// Shim that adapts the local storage implementation to the [Storage] interface.
//
// Translates Storage method calls into the local package's volume functions.
// The shim is stateless; the executor records all metadata in the deployment
// state file.
type storageLocalShim struct {
	root string // Base directory for all volumes.
}

// Returns a [Storage] backed by the local provider.
func newStorageLocal(root string) Storage {
	return &storageLocalShim{root: root}
}

// Creates a new volume directory on the host filesystem.
func (s *storageLocalShim) Create(ctx context.Context, opts VolumeOptions) (string, error) {
	id := fmt.Sprintf("%s%s", local.VolumeIDPrefix, crypto.RandHex(idLen))
	if _, err := local.CreateVolume(s.root, id, os.FileMode(opts.Mode)); err != nil {
		return "", crex.Wrap(ErrStorage, err)
	}
	return id, nil
}

// Removes a volume directory and its contents.
func (s *storageLocalShim) Remove(ctx context.Context, id string) error {
	if err := local.RemoveVolume(s.root, id); err != nil {
		return crex.Wrap(ErrStorage, err)
	}
	return nil
}

// No-op on the local provider.
//
// Local volumes are directories on the same machine as the containers and do
// not require device-level attachment. The executor binds the volume path
// into the container's OCI spec directly.
func (s *storageLocalShim) Attach(ctx context.Context, volume, instance string) error {
	return nil
}

// No-op on the local provider.
func (s *storageLocalShim) Detach(ctx context.Context, volume string) error {
	return nil
}

// Lists volume directories present on the filesystem.
func (s *storageLocalShim) List(ctx context.Context) ([]Volume, error) {
	ids, err := local.ListVolumes(s.root)
	if err != nil {
		return nil, crex.Wrap(ErrStorage, err)
	}
	out := make([]Volume, len(ids))
	for i, id := range ids {
		out[i] = Volume{ID: id}
	}
	return out, nil
}
