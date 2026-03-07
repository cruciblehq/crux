package provider

import "context"

// Interface for provider implementations.
//
// A provider provisions cruxd runtime instances. How the runtime is hosted
// is an implementation detail: the local provider starts a cruxd process on
// Linux and a Lima VM on macOS; a remote provider could create a cloud
// compute instance. Lifecycle methods ([Provision], [Deprovision], [Start],
// and [Stop]) block until the runtime reaches the expected target state. If
// it does not converge, the provider must revert any partial changes and
// return an error. Long-running operations must support context cancellation.
// When cancelled, the provider should stop in-flight work and revert changes.
type Backend interface {

	// Provisions a cruxd runtime instance.
	//
	// If a runtime with the configured name is already running, it is reused.
	// If provisioning fails or the runtime does not reach [StateRunning], the
	// provider tears down any partial state and returns an error.
	Provision(ctx context.Context, config *Config) error

	// Tears down a cruxd instance and all its persistent state.
	//
	// If the instance cannot be fully removed, partial artifacts are cleaned
	// up on a best-effort basis and an error is returned.
	Deprovision(ctx context.Context, name string) error

	// Resumes a stopped instance and blocks until it is reachable.
	//
	// If the instance does not reach [StateRunning], the provider stops any
	// partially-started resources and returns an error.
	Start(ctx context.Context, name string) error

	// Stops a running instance without deprovisioning it.
	//
	// Returns an error if the instance does not reach [StateStopped].
	Stop(ctx context.Context, name string) error

	// Returns the current lifecycle state of an instance.
	//
	// Must reflect the actual state of the underlying runtime. If the name
	// has never been provisioned, returns [StateNotProvisioned].
	Status(ctx context.Context, name string) (State, error)

	// Runs a command on the instance's host.
	Exec(ctx context.Context, name string, command string, args ...string) (*ExecResult, error)

	// Returns a client connected to the cruxd instance.
	Client(ctx context.Context, name string) (Client, error)
}
