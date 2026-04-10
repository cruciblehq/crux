package subsystem

// Identifies a subsystem domain.
//
// Each constant names a domain that can appear after the dot prefix in a
// grant expression (e.g. ".seccomp", ".file"). The domain selects which
// subsystem resolves and applies the grant.
type Domain string

const (
	DomainSeccomp Domain = "seccomp"  // Syscall filter.
	DomainFile    Domain = "file"     // LSM file access.
	DomainNet     Domain = "net"      // LSM network access.
	DomainMACCap  Domain = "mac_cap"  // LSM capability mediation.
	DomainSignal  Domain = "signal"   // LSM signal delivery.
	DomainExec    Domain = "exec"     // LSM exec mediation.
	DomainUnix    Domain = "unix"     // LSM unix socket access.
	DomainDBus    Domain = "dbus"     // LSM D-Bus mediation.
	DomainMount   Domain = "mount"    // LSM mount mediation.
	DomainPtrace  Domain = "ptrace"   // LSM ptrace mediation.
	DomainIOUring Domain = "io_uring" // LSM io_uring access.
	DomainUserns  Domain = "userns"   // LSM user namespace creation.
	DomainCap     Domain = "cap"      // Linux capabilities.
	DomainFcap    Domain = "fcap"     // File capabilities.
	DomainRlimit  Domain = "rlimit"   // POSIX resource limits.
	DomainCgroup  Domain = "cgroup"   // Cgroup v2 controllers.
	DomainExpose  Domain = "expose"   // Virtual filesystem unmask.
)
