package lima

// Lifecycle state of the Lima VM.
type State int

const (
	StateNotProvisioned State = iota // Instance has not been provisioned.
	StateRunning                     // Instance is running and reachable.
	StateStopped                     // Instance exists but is not running.
)
