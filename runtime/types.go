package runtime

// Current state of the container runtime environment.
type State int

const (
	StateNotCreated State = iota // Runtime has not been provisioned.
	StateStopped                 // Runtime exists but is not running.
	StateRunning                 // Runtime is running and reachable.
)

// Human-readable representation of the state.
func (s State) String() string {
	switch s {
	case StateNotCreated:
		return "not created"
	case StateStopped:
		return "stopped"
	case StateRunning:
		return "running"
	default:
		return "unknown"
	}
}

// Output captured from a command executed inside the runtime.
type ExecResult struct {
	Stdout   string // Standard output from the command.
	Stderr   string // Standard error from the command.
	ExitCode int    // Process exit code (0 = success).
}
