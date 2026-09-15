package compute

import (
	"context"
	"io"

	"github.com/cruciblehq/spec/affordance/kernel"
)

// Resource requirements for provisioning a compute instance.
//
// Each provider maps these requirements to its native compute class. The local
// provider uses fixed resource allocation and ignores the sizing fields.
type ComputeOptions struct {
	CPUs   int         `codec:"cpus,omitempty"`   // Minimum virtual CPUs required; zero means no minimum.
	Memory int         `codec:"memory,omitempty"` // Minimum memory in GiB required; zero means no minimum.
	Disk   int         `codec:"disk,omitempty"`   // Minimum disk size in GiB required; zero means no minimum.
	Kernel kernel.Spec `codec:"kernel,omitempty"` // Kernel requirements applied at provisioning time.
}

// Manages the lifecycle of compute instances.
//
// A compute implementation provisions and operates instances on a specific
// infrastructure provider. The local implementation uses Lima to manage a
// VM on the developer's machine; cloud implementations manage instances in
// their respective platforms. All lifecycle methods block until the instance
// reaches the expected state. If the state does not converge, any partial
// changes are reverted and an error is returned. All long-running operations
// support context cancellation; when cancelled, the implementation stops
// in-flight operations and reverts any partial state. Instances are addressed
// by provider-assigned identifiers returned at provisioning time. The
// implementation must ensure that identifiers are unique and immutable, and
// that operations on a given identifier affect the same underlying instance.
type Compute interface {

	// Uploads a disk image to the provider and returns an opaque image ID.
	//
	// r must contain a valid disk image in the format expected by the provider.
	// The local provider expects a qcow2 image; other providers have their own
	// requirements. The returned image ID is provider-specific and is passed to
	// Provision when creating an instance from the image. The provider may
	// perform provider-specific validation checks on the image and return an
	// error if the checks fail.
	UploadImage(ctx context.Context, r io.Reader) (string, error)

	// Removes a previously uploaded disk image from the provider.
	//
	// id is the opaque identifier returned by [Compute.UploadImage]. After
	// removal, the image cannot be used to provision new instances. Instances
	// already provisioned from the image are not affected. Returns an error
	// if the image does not exist or if removal fails.
	RemoveImage(ctx context.Context, id string) error

	// Provisions a new instance from a previously uploaded image.
	//
	// img is the opaque identifier returned by [Compute.UploadImage]. opts
	// specifies the resource requirements for the instance; the provider maps
	// them to its native compute class. The same image can be used to provision
	// multiple instances by calling this method multiple times. Returns a
	// provider-assigned identifier for the new instance. Returns an error if
	// provisioning fails; in that case no partial state is left behind.
	Provision(ctx context.Context, img string, opts ComputeOptions) (string, error)

	// Tears down the instance and removes all associated persistent state.
	//
	// If the instance is running it is stopped first. All disk images,
	// snapshots, and provider-level resources are permanently deleted. This
	// operation is irreversible, so the instance cannot be recovered after
	// Deprovision returns. Returns an error if the instance has not been
	// provisioned or if teardown does not complete within the deadline
	// imposed by ctx; in both cases no partial state is left behind.
	Deprovision(ctx context.Context, id string) error

	// Starts the instance and blocks until it is reachable.
	//
	// The instance is considered reachable when it is fully booted and
	// accepting connections through containerd. Returns nil immediately if the
	// instance is already running. Returns an error if the instance has not
	// been provisioned or cannot be started within the deadline imposed by ctx.
	// In all error cases, the instance is left in the state it was in before
	// Start was called.
	Start(ctx context.Context, id string) error

	// Stops the instance without removing its persistent state.
	//
	// Sends a shutdown signal and blocks until the instance has halted.
	// Workloads running on the instance are stopped as part of the shutdown
	// sequence. Disk snapshots and all other persistent state are preserved,
	// so the instance can be resumed later with Start. Returns nil if the
	// instance is already stopped and an error if the instance has not been
	// provisioned or started, or if the shutdown does not complete within the
	// deadline imposed by ctx. In case of error, clean up is performed on a
	// best-effort basis but the instance may be left in a partially stopped
	// state.
	Stop(ctx context.Context, id string) error

	// Returns the current state of the instance.
	//
	// Returns [StateNotProvisioned] if the instance has never been provisioned
	// or has been successfully deprovisioned. Returns [StateStopped] if the
	// instance exists but is not running. Returns [StateRunning] if the
	// instance is up and reachable. Returns an error only if the provider
	// cannot be queried; an instance that exists in an indeterminate state is
	// reported via the State value, not an error.
	Status(ctx context.Context, id string) (State, error)

	// Returns the identifiers of all instances known to the provider.
	//
	// Includes instances in any state. Returns an empty slice when no
	// instances have been provisioned.
	List(ctx context.Context) ([]string, error)

	// Runs a command on the instance.
	//
	// Executes the command on a running instance, outside any container, while
	// streaming stdout and stderr as the output is produced. Returns the exit
	// code and a nil error when the process exits normally, and a non-nil error
	// only if the command could not be started or the context was cancelled
	// before the process completed.
	Exec(ctx context.Context, id string, stdout, stderr io.Writer, command string, args ...string) (int, error)

	// Sends an uncompressed tar archive to the instance and extracts it.
	//
	// Entries are applied as absolute paths on the instance filesystem,
	// preserving permissions, ownership, and timestamps, and changes persist
	// across Stop and Start cycles. Returns an error if the instance is not
	// running, if ctx is cancelled before the transfer completes, or if
	// extraction fails.
	Copy(ctx context.Context, id string, r io.Reader) error

	// Opens a client connection to the container runtime on the instance.
	//
	// The instance must be running, otherwise an error is returned. The
	// returned [Client] owns the underlying containerd connection and must be
	// closed when no longer needed. Multiple clients can be open against the
	// same instance concurrently.
	Connect(ctx context.Context, id string) (*Client, error)
}
