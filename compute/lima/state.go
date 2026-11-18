package lima

// Lifecycle state of the Lima VM (without regard to cruxd inside it).
type State int

const (
	StateNotProvisioned State = iota // VM does not exist.
	StateRunning                     // VM is running.
	StateStopped                     // VM exists but is not running.
)
