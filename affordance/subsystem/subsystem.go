package subsystem

import (
	"github.com/cruciblehq/crux/affordance/agl"
)

// Identifier of a runtime subsystem.
//
// Carried by parsed grants and used by the orchestrator to dispatch a grant
// to the correct subsystem implementation. The string values match the leading
// token of a grant expression (e.g. ".cap").
type Name string

const (
	NameCap       Name = "cap"       // Linux capabilities.
	NameFcap      Name = "fcap"      // File capabilities.
	NameMAC       Name = "mac"       // Mandatory access control.
	NameSeccomp   Name = "seccomp"   // Seccomp syscall filter.
	NameRlimit    Name = "rlimit"    // POSIX resource limits.
	NameCgroup    Name = "cgroup"    // Cgroup v2 controls.
	NameProvision Name = "provision" // Resource allocation declarations.
	NameNet       Name = "net"       // Container network spec.
	NameMount     Name = "mount"     // Kernel VFS filesystem mounts.
	NameVolume    Name = "volume"    // Persistent storage volumes.
	NameDevice    Name = "device"    // Container device nodes.
	NameKernel    Name = "kernel"    // VM kernel feature requirements.
)

// Contract every subsystem implements.
//
// A subsystem owns one slice of the Spec, wired at construction time. Build
// folds a single parsed grant into the wired-in slice.
type Subsystem interface {

	// Returns the subsystem identifier used for dispatch.
	//
	// The string value must match the leading token of a grant expression (e.g.
	// ".cap") so that the orchestrator can route to the correct subsystem.
	Name() Name

	// Folds a single parsed grant into the wired-in section.
	//
	// Subsystems must check the grant's syntax and semantics and return an
	// error if the grant is invalid or cannot be applied. Subsystems decide
	//  whether a repeated grant is a no-op, a merge, or a conflict.
	Build(g *agl.Model) error
}
