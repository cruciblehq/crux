package subsystem

import "github.com/cruciblehq/crex"

// IO scheduling priority class.
type ioPrioClass string

const (
	ioPrioRT   ioPrioClass = "rt"   // Real-time (highest priority, deadline scheduling).
	ioPrioBE   ioPrioClass = "be"   // Best-effort (default class, weight-based scheduling).
	ioPrioIdle ioPrioClass = "idle" // Idle (served only when no other I/O is pending).
)

// Partition mode for cpuset CPU isolation.
type cgroupPartition string

const (
	cgroupPartitionMember   cgroupPartition = "member"   // Non-isolated, shares parent's CPUs.
	cgroupPartitionRoot     cgroupPartition = "root"     // Partition root, owns its CPUs exclusively.
	cgroupPartitionIsolated cgroupPartition = "isolated" // Like root, but also removed from the scheduler's load-balancing.
)

// Node type within the cgroup hierarchy.
type cgroupNodeType string

const (
	cgroupNodeDomain   cgroupNodeType = "domain"   // Process-granularity cgroup (default).
	cgroupNodeThreaded cgroupNodeType = "threaded" // Thread-granularity cgroup.
)

// Device type.
type deviceKind string

const (
	deviceChar  deviceKind = "c" // Character device.
	deviceBlock deviceKind = "b" // Block device.
	deviceAll   deviceKind = "a" // All device types (wildcard).
)

// Converts a string to a deviceKind, returning an error for unknown values.
func parseDeviceKind(s string) (deviceKind, error) {
	k := deviceKind(s)
	switch k {
	case deviceChar, deviceBlock, deviceAll:
		return k, nil
	default:
		return "", crex.Wrapf(ErrGrantExpression, "unknown type %q", s)
	}
}

// Device access permission flag.
type deviceAccess rune

const (
	deviceRead  deviceAccess = 'r' // Read access.
	deviceWrite deviceAccess = 'w' // Write access.
	deviceMknod deviceAccess = 'm' // Mknod access.
)

// PSI trigger kind.
type psiKind string

const (
	psiSome psiKind = "some" // At least one task stalled.
	psiFull psiKind = "full" // All tasks stalled simultaneously.
)

// Converts a string to a psiKind, returning an error for unknown values.
func parsePsiKind(s string) (psiKind, error) {
	k := psiKind(s)
	switch k {
	case psiSome, psiFull:
		return k, nil
	default:
		return "", crex.Wrapf(ErrGrantExpression, "unknown kind %q", s)
	}
}
