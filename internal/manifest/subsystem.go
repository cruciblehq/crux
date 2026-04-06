package manifest

// Identifies a subsystem in the grant expression grammar.
//
// Each constant names a subsystem recognized in the expression grammar.
// [SubRef] is a special case: it is the implicit subsystem assigned to
// grants that use bare names instead of a dot-prefixed selector (e.g.
// "fd/dup" decodes as SubRef with Expr "fd/dup"). Ref grants are resolved
// during affordance building.
type Subsystem string

const (
	SubRef     Subsystem = "ref"      // Affordance reference.
	SubSeccomp Subsystem = "seccomp"  // Syscall filter.
	SubFile    Subsystem = "file"     // LSM file access.
	SubNet     Subsystem = "net"      // LSM network access.
	SubMACCap  Subsystem = "mac_cap"  // LSM capability mediation.
	SubSignal  Subsystem = "signal"   // LSM signal delivery.
	SubExec    Subsystem = "exec"     // LSM exec mediation.
	SubUnix    Subsystem = "unix"     // LSM unix socket access.
	SubDBus    Subsystem = "dbus"     // LSM D-Bus mediation.
	SubMount   Subsystem = "mount"    // LSM mount mediation.
	SubPtrace  Subsystem = "ptrace"   // LSM ptrace mediation.
	SubIOUring Subsystem = "io_uring" // LSM io_uring access.
	SubUserns  Subsystem = "userns"   // LSM user namespace creation.
	SubCap     Subsystem = "cap"      // Linux capabilities.
	SubFcap    Subsystem = "fcap"     // File capabilities.
	SubRlimit  Subsystem = "rlimit"   // POSIX resource limits.
	SubCgroup  Subsystem = "cgroup"   // Cgroup v2 controllers.
	SubExpose  Subsystem = "expose"   // Virtual filesystem unmask.
)

// Parsing extension point for the sandbox expression grammar.
//
// Each implementor handles one or more subsystem prefixes and translates
// the grant into mutations on its own struct fields. subsystemRouter maps
// every Subsystem constant to exactly one subsystemHandler, so adding a
// new subsystem only requires wiring the constant and implementing this
// interface on the target type.
type subsystemHandler interface {

	// Applies a single grant to the receiver.
	//
	// subsystem is the subsystem constant that was used to route to this
	// implementor. A single type may handle several subsystems and uses
	// subsystem to distinguish them internally. rest is the expression
	// payload from the grant. args carries the structured map form when
	// the YAML grant uses a single-key map whose value is a list of strings.
	// Each string is a "key [value]" pair. For bare string grants args is nil.
	applyGrant(subsystem Subsystem, rest string, args []string) error
}
