package shared

import (
	"github.com/cruciblehq/crux/internal/manifest/grant"
)

// Identifier of a runtime subsystem.
//
// Carried by parsed grants and used by the orchestrator to dispatch a grant
// to the correct subsystem implementation. The string values match the leading
// token of a grant expression (e.g. ".cap").
type Name string

const (
	NameCap     Name = "cap"     // Linux capabilities.
	NameFcap    Name = "fcap"    // File capabilities.
	NameMAC     Name = "mac"     // Mandatory access control.
	NameSeccomp Name = "seccomp" // Seccomp syscall filter.
	NameRlimit  Name = "rlimit"  // POSIX resource limits.
	NameCgroup  Name = "cgroup"  // Cgroup v2 controls.
)

// Contract every subsystem implements.
//
// A subsystem owns one slice of the [Spec], wired at construction time. Build
// folds a single parsed grant into the wired-in slice. Merge folds the matching
// slice of another whole [Spec] into the wired-in slice.
type Subsystem interface {

	// Returns the subsystem identifier used for dispatch.
	//
	// The string value must match the leading token of a grant expression (e.g.
	// ".cap") so that the orchestrator can route to the correct subsystem.
	Name() Name

	// Folds a single parsed grant into the wired-in section.
	//
	// Subsystems must check the grant's syntax and semantics and return an
	// error if the grant is invalid or cannot be applied.
	Build(g grant.Grant) error

	// Folds the matching section of src into the wired-in section.
	//
	// Subsystems check any relevant sections of src and return an error if the
	// merge cannot be performed. The merge logic is subsystem-specific and may
	// be non-commutative. For example, a subsystem might reject a merge if src
	// contains entries that conflict with their internal state.
	Merge(src Spec) error
}
