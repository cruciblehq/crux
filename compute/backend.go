package compute

// Entry point for a provider's infrastructure capabilities.
//
// A Backend groups the three resource managers that every provider must
// implement: [Compute] for instance lifecycle, [Storage] for persistent
// volumes, and [Network] for topology and firewall rules. Callers obtain
// a Backend via [BackendFor] and then access the manager they need.
type Backend interface {

	// Returns the compute instance manager for this provider.
	Compute() Compute

	// Returns the persistent volume manager for this provider.
	Storage() Storage

	// Returns the network topology and firewall manager for this provider.
	Network() Network
}
