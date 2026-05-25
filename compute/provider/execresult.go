package provider

// Output captured from a command executed on the instance's host.
type ExecResult struct {
	Stdout   string // Standard output from the command.
	Stderr   string // Standard error from the command.
	ExitCode int    // Process exit code (0 = success).
}

// Returns a new [ExecResult].
func NewExecResult(stdout, stderr string, exitCode int) *ExecResult {
	return &ExecResult{Stdout: stdout, Stderr: stderr, ExitCode: exitCode}
}
