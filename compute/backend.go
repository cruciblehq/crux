package compute

import (
	"context"
	"io"
)

// Interface for provider backend implementations.
//
// A provider manages the lifecycle of compute hosts. The local provider uses
// Lima to manage a VM on the local machine; each cloud provider port manages
// hosts in their respective platforms. All lifecycle methods block until the
// host reaches the expected state. If the state does not converge, any partial
// changes are reverted and an error is returned. All long-running operations
// support context cancellation; when cancelled, the provider stops in-flight
// operations and reverts any partial state. Hosts are addressed by name. The
// provider assigns a unique name to each host at provisioning time and uses
// that name to identify the host in all subsequent operations. The provider
// must ensure that names are unique and immutable, and that operations on a
// given name affect the same underlying host.
type Backend interface {

	// Uploads a disk image to the provider and returns an opaque image ID.
	//
	// r must contain a valid disk image in the format expected by the provider.
	// The local provider expects a qcow2 image; other providers have their own
	// requirements. The returned image ID is provider-specific and is passed to
	// Provision when creating a host from the image. The provider may perform
	// provider-specific validation checks on the image and return an error if
	// the checks fail.
	Upload(ctx context.Context, r io.Reader) (string, error)

	// Provisions a new host from a previously uploaded image.
	//
	// img is the opaque identifier returned by [Backend.Upload]. name is an
	// arbitrary string that permanently identifies the host; all subsequent
	// operations on the host pass the same name. opts specifies the resource
	// requirements for the host; the backend maps them to its native compute
	// class. The same image can be used to provision multiple hosts by calling
	// this method with different names. An error is returned if a host with
	// the given name already exists or if provisioning fails.
	Provision(ctx context.Context, img, name string, opts Options) error

	// Tears down the named host and removes all associated persistent state.
	//
	// If the host is running it is stopped first. All disk images, snapshots,
	// and provider-level resources are permanently deleted. This operation is
	// irreversible, so the host cannot be recovered after Deprovision returns.
	// Returns an error if the host has not been provisioned or if teardown does
	// not complete within the deadline imposed by ctx; in both cases no partial
	// state is left behind.
	Deprovision(ctx context.Context, name string) error

	// Starts the named host and blocks until it is reachable.
	//
	// The host is considered reachable when it is fully booted and accepting
	// connections through containerd. Returns nil immediately if the host is
	// already running. Returns an error if the host has not been provisioned
	// or cannot be started within the deadline imposed by ctx. In all error
	// cases, the host is left in the state it was in before Start was called.
	Start(ctx context.Context, name string) error

	// Stops the named host without removing its persistent state.
	//
	// Sends a shutdown signal and blocks until the host has halted. Workloads
	// running on the host are stopped as part of the shutdown sequence. Disk
	// snapshots and all other persistent state are preserved, so the host can
	// be resumed later with Start. Returns nil if the host is already stopped
	// and an error if the host has not been provisioned or started, or if the
	// shutdown does not complete within the deadline imposed by ctx. In case
	// of error, clean up is performed on a best-effort basis but the host may
	// be left in a partially stopped state.
	Stop(ctx context.Context, name string) error

	// Returns the current state of the named host.
	//
	// Returns [StateNotProvisioned] if the host has never been provisioned or
	// has been successfully deprovisioned. Returns [StateStopped] if the host
	// exists but is not running. Returns [StateRunning] if the host is up and
	// reachable. Returns an error only if the provider cannot be queried; a
	// host that exists in an indeterminate state is reported via the State
	// value, not an error. The host's state is not altered.
	Status(ctx context.Context, name string) (State, error)

	// Returns the names of all hosts known to the provider.
	//
	// Includes hosts in any state. Names are returned in lexicographic order.
	// Returns an empty slice when no hosts have been provisioned.
	List(ctx context.Context) ([]string, error)

	// Runs a command on the named host.
	//
	// Executes the command on a host with state [StateRunning], outside any
	// container, while streaming stdout and stderr as the output is produced.
	// Returns the exit code and a nil error when the process exits normally,
	// and a non-nil error only if the command could not be started or the
	// context was cancelled before the process completed. The host's state
	// is not altered.
	Exec(ctx context.Context, name string, stdout, stderr io.Writer, command string, args ...string) (int, error)

	// Sends an uncompressed tar archive to the named host and extracts it.
	//
	// Entries are applied as absolute paths on the host filesystem, preserving
	// permissions, ownership, and timestamps, and changes persist across Stop
	// and Start cycles. Returns an error if the host is not running, if ctx is
	// cancelled before the transfer completes, or if extraction fails.
	Copy(ctx context.Context, name string, r io.Reader) error

	// Opens a client connection to the container runtime on the named host.
	//
	// The host must be running, otherwise an error is returned. The returned
	// [Client] owns the underlying containerd connection and must be closed
	// when no longer needed. Multiple clients can be open against the same
	// host concurrently.
	Connect(ctx context.Context, name string) (*Client, error)
}
