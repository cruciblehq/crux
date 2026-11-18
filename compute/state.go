package compute

// Lifecycle state of a machine.
type State int

const (
	StateNotProvisioned State = iota // Machine has not been provisioned.
	StateRunning                     // Machine is running and cruxd answers.
	StateStopped                     // Machine exists but is not running.
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
