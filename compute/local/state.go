package local

// Lifecycle state of the local compute instance.
type State int

const (
	StateNotProvisioned State = iota // Instance has not been provisioned.
	StateRunning                     // Instance is running and reachable.
	StateStopped                     // Instance exists but is not running.
)
