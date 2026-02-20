package runtime

// Output captured from a command executed inside the runtime.
type ExecResult struct {
	Stdout   string // Standard output from the command.
	Stderr   string // Standard error from the command.
	ExitCode int    // Process exit code (0 = success).
}
