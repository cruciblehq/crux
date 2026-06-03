package compute

// Compute backend implementations types.
//
// Each value corresponds to an infrastructure provider and each provider has
// a registered [provider.Backend] that can be retrieved with [BackendFor].
type Provider string

const (

	// A VM running on the local machine.
	//
	// The local provider uses a VM to run containers on the local machine. This
	// allows the local environment to better resemble production environments,
	// while still providing a simple setup with no external dependencies. The
	// local backend provisions the VM and schedules work there by connecting
	// to the containerd socket exposed by the VM. Intended for development.
	Local Provider = "local"
)
