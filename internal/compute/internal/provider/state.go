package provider

// Lifecycle state of a cruxd instance.
type State int

const (
	StateNotProvisioned State = iota // Instance has not been provisioned.
	StateRunning                     // Instance is running and reachable.
	StateStopped                     // Instance exists but is not running.
)

// Canonical string representation of the state.
func (s State) String() string {
	switch s {
	case StateNotProvisioned:
		return "not provisioned"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}
