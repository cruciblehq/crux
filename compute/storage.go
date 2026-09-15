package compute

import "context"

// Configuration for creating a volume.
type VolumeOptions struct {
	Size int `codec:"size,omitempty"` // Capacity in MiB; zero means provider default.
	Mode int `codec:"mode,omitempty"` // Unix permission mode; zero defaults to 0o700 for the local provider.
}

// Describes an existing volume.
type Volume struct {
	ID       string `codec:"id"`                 // Provider-assigned identifier.
	Size     int    `codec:"size,omitempty"`     // Allocated capacity in MiB.
	Instance string `codec:"instance,omitempty"` // Identifier of the attached instance; empty if detached.
}

// Manages the lifecycle of persistent storage volumes.
//
// Volumes exist independently of compute instances and are attached or
// detached as needed. The provider assigns a unique identifier to each volume
// at creation time. All identifiers are provider-specific and opaque to the
// caller; the executor persists them in the deployment state.
type Storage interface {

	// Creates a new volume and returns its provider-assigned identifier.
	//
	// The volume exists independently of any instance until explicitly
	// attached. Returns an error if the provider cannot allocate the
	// requested capacity.
	Create(ctx context.Context, opts VolumeOptions) (string, error)

	// Permanently removes the volume and its data.
	//
	// The volume must not be attached to an instance. Returns an error if the
	// volume does not exist or is still attached.
	Remove(ctx context.Context, id string) error

	// Attaches a volume to a running instance.
	//
	// volume is the identifier returned by [Storage.Create]. instance is the
	// identifier returned by [Compute.Provision]. The provider makes the
	// volume's backing store accessible on the instance. The mechanism is
	// provider-specific: the local provider bind-mounts a directory; cloud
	// providers attach a block device. Returns an error if either identifier
	// is invalid or if the volume is already attached.
	Attach(ctx context.Context, volume, instance string) error

	// Detaches a volume from its instance.
	//
	// The volume's data is preserved. Returns an error if the volume is not
	// currently attached.
	Detach(ctx context.Context, volume string) error

	// Returns all volumes known to the provider.
	//
	// Includes both attached and detached volumes. Returns an empty slice
	// when no volumes have been created.
	List(ctx context.Context) ([]Volume, error)
}
