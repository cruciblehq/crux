package manifest

import "github.com/cruciblehq/crux/internal/subsystem"

// Domain identifies a subsystem domain in the grant expression grammar.
type Domain = subsystem.Domain

const (
	DomainRef     = subsystem.DomainRef     // Affordance reference.
	DomainSeccomp = subsystem.DomainSeccomp // Syscall filter.
	DomainFile    = subsystem.DomainFile    // LSM file access.
	DomainNet     = subsystem.DomainNet     // LSM network access.
	DomainMACCap  = subsystem.DomainMACCap  // LSM capability mediation.
	DomainSignal  = subsystem.DomainSignal  // LSM signal delivery.
	DomainExec    = subsystem.DomainExec    // LSM exec mediation.
	DomainUnix    = subsystem.DomainUnix    // LSM unix socket access.
	DomainDBus    = subsystem.DomainDBus    // LSM D-Bus mediation.
	DomainMount   = subsystem.DomainMount   // LSM mount mediation.
	DomainPtrace  = subsystem.DomainPtrace  // LSM ptrace mediation.
	DomainIOUring = subsystem.DomainIOUring // LSM io_uring access.
	DomainUserns  = subsystem.DomainUserns  // LSM user namespace creation.
	DomainCap     = subsystem.DomainCap     // Linux capabilities.
	DomainFcap    = subsystem.DomainFcap    // File capabilities.
	DomainRlimit  = subsystem.DomainRlimit  // POSIX resource limits.
	DomainCgroup  = subsystem.DomainCgroup  // Cgroup v2 controllers.
	DomainExpose  = subsystem.DomainExpose  // Virtual filesystem unmask.
)
