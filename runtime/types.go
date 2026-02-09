package runtime

// Current state of the container runtime environment.
type Status int

const (
	StatusNotCreated Status = iota // Runtime has not been provisioned.
	StatusStopped                  // Runtime exists but is not running.
	StatusRunning                  // Runtime is running and reachable.
)

// Human-readable representation of the status.
func (s Status) String() string {
	switch s {
	case StatusNotCreated:
		return "not created"
	case StatusStopped:
		return "stopped"
	case StatusRunning:
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
