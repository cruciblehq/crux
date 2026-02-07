package vm

// Represents the current state of the virtual machine.
type Status int

const (
	StatusNotCreated Status = iota // VM does not exist.
	StatusStopped                  // VM exists but is not running.
	StatusRunning                  // VM is running and reachable.
)

// Returns a human-readable representation of the VM status.
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

// Result of executing a command inside the VM.
type ExecResult struct {
	Stdout   string // Standard output from the command.
	Stderr   string // Standard error from the command.
	ExitCode int    // Process exit code (0 = success).
}
