package runtime

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// Options for executing a command inside a container.
//
// When Env is set it replaces the process environment entirely. When Workdir
// is set it overrides the process working directory. A zero-value ExecOptions
// inherits everything from the container's OCI spec.
type ExecOptions struct {
	Env     []string // Environment as KEY=VAL pairs, replaces default if set.
	Workdir string   // Working directory override.
}

// Output captured from a command executed inside a container.
type ExecResult struct {
	Stdout   string // Standard output from the command.
	Stderr   string // Standard error from the command.
	ExitCode int    // Process exit code (0 = success).
}

// Sequence counter for generating unique exec process identifiers.
var execSeq uint64

// Returns a unique exec process identifier.
func nextExecID() string {
	return fmt.Sprintf("exec-%d", atomic.AddUint64(&execSeq, 1))
}

// Merges override env vars on top of a base env slice.
//
// Both base and overrides use KEY=VAL format. Overrides with the same key
// replace the base entry; new keys are appended. Order is preserved for
// base entries, overrides appear at their original position or at the end.
func mergeEnv(base, overrides []string) []string {
	merged := make(map[string]string, len(base)+len(overrides))
	for _, entry := range base {
		if k, v, ok := strings.Cut(entry, "="); ok {
			merged[k] = v
		}
	}
	for _, entry := range overrides {
		if k, v, ok := strings.Cut(entry, "="); ok {
			merged[k] = v
		}
	}

	result := make([]string, 0, len(merged))
	for k, v := range merged {
		result = append(result, k+"="+v)
	}
	return result
}
