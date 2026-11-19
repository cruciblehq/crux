package compute

import (
	"context"
	"io"

	"github.com/cruciblehq/spec/agent"
)

// Name of an infrastructure provider, used to look up its [Provider].
type ProviderName string

const (

	// A VM on this computer, run by Lima.
	//
	// The local provider lets the local environment resemble production with
	// no external dependencies. crux talks to cruxd inside the VM over mutual
	// TLS. Intended for development.
	Local ProviderName = "local"
)

// Resource requirements for provisioning a machine.
//
// Each provider maps these requirements to its native machine class. The local
// provider uses fixed resources and ignores the sizing fields.
type ProvisionOptions struct {
	CPUs   int `codec:"cpus,omitempty"`   // Minimum virtual CPUs required; zero means no minimum.
	Memory int `codec:"memory,omitempty"` // Minimum memory in GiB required; zero means no minimum.
	Disk   int `codec:"disk,omitempty"`   // Minimum disk size in GiB required; zero means no minimum.
}

// Creates and runs machines on one infrastructure provider.
//
// The local provider runs a single Lima VM on this computer. Cloud providers
// would manage machines on their platforms. Every lifecycle method blocks
// until the machine reaches the expected state. If it does not get there,
// partial changes are reverted and an error is returned. Cancelling ctx stops
// the operation in flight and reverts partial state. Machines are addressed
// by the identifiers that Provision returns, which are unique and never change.
type Provider interface {

	// Uploads a disk image to the provider and returns an opaque image ID.
	//
	// r must hold a disk image in the format the provider expects. The local
	// provider expects a qcow2 image. The returned ID is passed to Provision.
	// The provider may check the image and return an error if the checks fail.
	UploadImage(ctx context.Context, r io.Reader) (string, error)

	// Removes a disk image uploaded with [Provider.UploadImage].
	//
	// After removal, the image cannot be used to provision new machines.
	// Machines already provisioned from it are not affected.
	RemoveImage(ctx context.Context, id string) error

	// Provisions a new machine from an uploaded image.
	//
	// img is an ID returned by [Provider.UploadImage]. opts gives the resource
	// requirements, which the provider maps to its native machine class. The
	// same image can provision many machines. Returns the new machine's
	// identifier. On failure, no partial state is left behind.
	Provision(ctx context.Context, img string, opts ProvisionOptions) (string, error)

	// Destroys the machine and all its persistent state.
	//
	// A running machine is stopped first. Disks, snapshots, and provider
	// resources are deleted, and the machine cannot be recovered. Returns an
	// error if the machine has not been provisioned or if teardown does not
	// finish before ctx ends.
	Deprovision(ctx context.Context, id string) error

	// Starts the machine and blocks until cruxd on it is reachable.
	//
	// Returns nil at once if the machine is already running. Returns an error
	// if the machine has not been provisioned or does not start before ctx
	// ends. On error, the machine is left in the state it was in before.
	Start(ctx context.Context, id string) error

	// Stops the machine and keeps its persistent state.
	//
	// Containers on the machine stop as part of the shutdown. The machine can
	// be started again with Start. Returns nil if it is already stopped, and an
	// error if it has not been provisioned or does not stop before ctx ends. On
	// error, cleanup is best effort and the machine may be partly stopped.
	Stop(ctx context.Context, id string) error

	// Returns the current state of the machine.
	//
	// Returns [StateNotProvisioned] if the machine was never provisioned or has
	// been deprovisioned, [StateStopped] if it exists but is not running, and
	// [StateRunning] if it is up and cruxd answers. Returns an error only if
	// the provider cannot be queried.
	Status(ctx context.Context, id string) (State, error)

	// Returns the identifiers of all machines the provider knows, in any state.
	List(ctx context.Context) ([]string, error)

	// Runs a command on the machine, outside any container.
	//
	// stdout and stderr receive output as it is produced. Returns the exit
	// code and a nil error when the process exits, whatever the code. Returns
	// an error only if the command could not start or ctx ended first.
	Exec(ctx context.Context, id string, stdout, stderr io.Writer, command string, args ...string) (int, error)

	// Opens a session with cruxd on the machine.
	//
	// The machine must be running. The returned agent owns its connection
	// and must be closed when no longer needed. Many sessions can be open
	// against the same machine at once.
	Connect(ctx context.Context, id string) (agent.Agent, error)
}
