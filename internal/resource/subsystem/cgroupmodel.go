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

// Resource limits for the container's Linux cgroup v2 hierarchy.
//
// Controls CPU, memory, I/O, process count, CPU affinity, huge pages, RDMA,
// miscellaneous scalar resources, pressure stall thresholds, and hierarchy
// structure. Each controller is a typed struct. Fields whose most restrictive
// value is not the Go zero value carry a codec default tag so the runtime
// applies the restrictive setting when no grant is present (e.g. cgroup.freeze
// defaults to true, cpu.weight to 1).
type cgroup struct {
	Core    cgroupCore      `codec:"core"`              // cgroup hierarchy and freeze controls.
	CPU     cgroupCPU       `codec:"cpu"`               // CPU time and scheduling weight.
	Memory  cgroupMemory    `codec:"memory"`            // Memory limits and protections.
	IO      cgroupIO        `codec:"io"`                // Block I/O weight, per-device limits, and cost model.
	Pids    cgroupPids      `codec:"pids"`              // Process count limit.
	CPUSet  cgroupCPUSet    `codec:"cpuset"`            // CPU and memory node affinity.
	HugeTLB []cgroupHugeTLB `codec:"hugetlb,omitempty"` // Huge page limits per page size.
	RDMA    []cgroupRDMA    `codec:"rdma,omitempty"`    // RDMA resource limits per device.
	Misc    []cgroupMisc    `codec:"misc,omitempty"`    // Miscellaneous scalar resource limits (e.g., SEV slots).
	Devices []cgroupDevice  `codec:"devices,omitempty"` // Device access permissions (BPF_PROG_TYPE_CGROUP_DEVICE).
	PSI     cgroupPSI       `codec:"psi"`               // Pressure stall information triggers.
}

// Core cgroup hierarchy controls.
//
// Manages the cgroup node itself rather than any specific resource: whether
// the cgroup is frozen, its type within the hierarchy, depth and descendant
// limits, and which controllers are delegated to children.
type cgroupCore struct {
	Freeze         bool           `codec:"freeze,default=true"           knob:"cgroup.freeze"`          // cgroup.freeze. Defaults to true (frozen).
	Type           cgroupNodeType `codec:"type,omitempty,default=domain" knob:"cgroup.type"`            // cgroup.type ("domain", "threaded").
	MaxDescendants uint32         `codec:"max_descendants,omitempty"     knob:"cgroup.max.descendants"` // cgroup.max.descendants. Zero means no children allowed.
	MaxDepth       uint32         `codec:"max_depth,omitempty"           knob:"cgroup.max.depth"`       // cgroup.max.depth. Zero means no nesting allowed.
	SubtreeControl []string       `codec:"subtree_control,omitempty"`                                   // cgroup.subtree_control (e.g., "cpu", "memory", "io", "pids").
}

// CPU bandwidth and scheduling priority.
//
// Combines hard bandwidth limiting (quota per period) with proportional weight
// scheduling among siblings. The controller also supports burst accumulation
// and an idle-priority mode.
type cgroupCPU struct {
	Max    uint64 `codec:"max"                             knob:"cpu.max"`        // cpu.max quota in microseconds per period. Zero means no CPU time.
	Period uint64 `codec:"period,omitempty,default=100000" knob:"cpu.max.period"` // cpu.max period in microseconds. Defaults to 100000 (100ms kernel default).
	Weight uint16 `codec:"weight,omitempty,default=1"      knob:"cpu.weight"`     // cpu.weight (1–10000). Minimum share.
	Burst  uint64 `codec:"burst,omitempty"                 knob:"cpu.max.burst"`  // cpu.max.burst in microseconds. Zero means no burst.
	Idle   bool   `codec:"idle,default=true"               knob:"cpu.idle"`       // cpu.idle. Defaults to true (idle-priority).
}

// Memory limits and reclaim protection.
//
// Provides a hard ceiling enforced by the OOM killer, a throttling threshold
// that triggers reclaim before the hard limit, and hard/soft reclaim floors.
// Separate limits cover swap and compressed swap (zswap). The OOM group flag
// controls whether the kernel kills all cgroup processes together when the
// limit is exceeded.
type cgroupMemory struct {
	Max      uint64 `codec:"max"                 knob:"memory.max"`       // memory.max in bytes. Zero means no memory.
	High     uint64 `codec:"high,omitempty"      knob:"memory.high"`      // memory.high in bytes. Zero means immediate throttling.
	Min      uint64 `codec:"min,omitempty"       knob:"memory.min"`       // memory.min in bytes (hard reclaim floor). Zero means no guarantee.
	Low      uint64 `codec:"low,omitempty"       knob:"memory.low"`       // memory.low in bytes (soft reclaim floor). Zero means no protection.
	SwapMax  uint64 `codec:"swap_max,omitempty"  knob:"memory.swap.max"`  // memory.swap.max in bytes. Zero means no swap.
	SwapHigh uint64 `codec:"swap_high,omitempty" knob:"memory.swap.high"` // memory.swap.high in bytes. Zero means immediate swap throttle.
	OOMGroup bool   `codec:"oom_group,omitempty" knob:"memory.oom.group"` // memory.oom.group. True kills all cgroup processes on OOM.
	ZswapMax uint64 `codec:"zswap_max,omitempty" knob:"memory.zswap.max"` // memory.zswap.max in bytes. Zero means no zswap.
}

// Block I/O weight, priority, and per-device limits.
//
// Combines global proportional weight scheduling with per-device bandwidth
// caps, latency targets, and blk-iocost model tuning. Devices are identified
// by major:minor number.
type cgroupIO struct {
	Weight    uint16            `codec:"weight,omitempty,default=1" knob:"io.weight"`     // io.weight (1–10000). Minimum share.
	PrioClass ioPrioClass       `codec:"prio_class,omitempty"       knob:"io.prio.class"` // io.prio.class ("rt", "be", "idle"). Empty means no override.
	Max       []cgroupIOMax     `codec:"max,omitempty"`                                   // Per-device bandwidth and IOPS limits.
	Latency   []cgroupIOLatency `codec:"latency,omitempty"`                               // Per-device latency targets.
	Cost      []cgroupIOCost    `codec:"cost,omitempty"`                                  // Per-device cost model coefficients.
	CostQoS   []cgroupIOCostQoS `codec:"cost_qos,omitempty"`                              // Per-device cost model QoS parameters.
}

// Per-device I/O bandwidth and IOPS caps.
//
// Corresponds to one line in the io.max interface file. Sets hard upper
// bounds on throughput and operation rate for a single block device.
type cgroupIOMax struct {
	Major uint32 `codec:"major"` // Device major number.
	Minor uint32 `codec:"minor"` // Device minor number.
	Rbps  uint64 `codec:"rbps"`  // Max read bytes/sec. Zero means no reads.
	Wbps  uint64 `codec:"wbps"`  // Max write bytes/sec. Zero means no writes.
	Riops uint64 `codec:"riops"` // Max read IOPS. Zero means no reads.
	Wiops uint64 `codec:"wiops"` // Max write IOPS. Zero means no writes.
}

// Per-device I/O latency target.
//
// Corresponds to one line in the io.latency interface file. The kernel
// throttles sibling cgroups to help this cgroup meet its target.
type cgroupIOLatency struct {
	Major  uint32 `codec:"major"`  // Device major number.
	Minor  uint32 `codec:"minor"`  // Device minor number.
	Target uint64 `codec:"target"` // Latency target in microseconds. Zero means no guarantee.
}

// Process and thread count limit.
//
// Caps the total number of tasks (processes and threads) in the cgroup,
// preventing fork bombs from exhausting the system PID space.
type cgroupPids struct {
	Max uint64 `codec:"max" knob:"pids.max"` // pids.max. Zero means no processes.
}

// CPU and memory node affinity.
//
// Restricts which CPUs and NUMA memory nodes the cgroup can use, and
// optionally promotes it to an exclusive scheduling partition.
type cgroupCPUSet struct {
	CPUs      string          `codec:"cpus,omitempty"                     knob:"cpuset.cpus"`           // cpuset.cpus (e.g., "0-3", "0,2,4"). Empty means no CPUs.
	Mems      string          `codec:"mems,omitempty"                     knob:"cpuset.mems"`           // cpuset.mems (e.g., "0", "0-1"). Empty means no memory nodes.
	Partition cgroupPartition `codec:"partition,omitempty,default=member" knob:"cpuset.cpus.partition"` // cpuset.cpus.partition ("member", "root", "isolated").
}

// Huge page limit for one page size.
//
// Caps huge page usage for a single page size (e.g. 2MB, 1GB), covering
// both on-demand and pre-reserved allocations.
type cgroupHugeTLB struct {
	Size    string `codec:"size"`               // Page size (e.g., "2MB", "1GB").
	Max     uint64 `codec:"max"`                // hugetlb.<size>.max in bytes. Zero means none.
	RsvdMax uint64 `codec:"rsvd_max,omitempty"` // hugetlb.<size>.rsvd.max in bytes. Zero means none.
}

// RDMA resource limit for one HCA device.
//
// Caps the number of HCA handles and objects that the cgroup can allocate
// on a single RDMA device.
type cgroupRDMA struct {
	Device    string `codec:"device"`               // HCA device name (e.g., "mlx5_0").
	HcaHandle uint32 `codec:"hca_handle,omitempty"` // Max HCA handles. Zero means none.
	HcaObject uint32 `codec:"hca_object,omitempty"` // Max HCA objects. Zero means none.
}

// Per-device I/O cost model coefficients.
//
// Corresponds to one line in the io.cost.model interface file. Provides
// measured device capacities so the blk-iocost controller can convert
// bytes and operations into a unified cost unit.
type cgroupIOCost struct {
	Major     uint32 `codec:"major"`     // Device major number.
	Minor     uint32 `codec:"minor"`     // Device minor number.
	Rbps      uint64 `codec:"rbps"`      // Sequential read bytes/sec capacity.
	Rseqiops  uint64 `codec:"rseqiops"`  // Sequential read IOPS capacity.
	Rrandiops uint64 `codec:"rrandiops"` // Random read IOPS capacity.
	Wbps      uint64 `codec:"wbps"`      // Sequential write bytes/sec capacity.
	Wseqiops  uint64 `codec:"wseqiops"`  // Sequential write IOPS capacity.
	Wrandiops uint64 `codec:"wrandiops"` // Random write IOPS capacity.
}

// Per-device I/O cost QoS parameters.
//
// Corresponds to one line in the io.cost.qos interface file. Tunes how
// aggressively the blk-iocost controller throttles I/O to meet latency
// targets on a single block device.
type cgroupIOCostQoS struct {
	Major uint32  `codec:"major"`          // Device major number.
	Minor uint32  `codec:"minor"`          // Device minor number.
	Rpct  float64 `codec:"rpct,omitempty"` // Read latency percentile (0.00–100.00).
	Rlat  uint64  `codec:"rlat,omitempty"` // Read latency target in microseconds.
	Wpct  float64 `codec:"wpct,omitempty"` // Write latency percentile (0.00–100.00).
	Wlat  uint64  `codec:"wlat,omitempty"` // Write latency target in microseconds.
	Min   float64 `codec:"min,omitempty"`  // Minimum weight fraction (0.00–1.00).
	Max   float64 `codec:"max,omitempty"`  // Maximum weight fraction (0.00–1.00).
}

// Miscellaneous scalar resource limit.
//
// Covers host resources that don't belong to a dedicated controller,
// such as AMD SEV encryption slots.
type cgroupMisc struct {
	Resource string `codec:"resource"` // Resource name (e.g., "sev", "sev_es").
	Max      uint64 `codec:"max"`      // Maximum count. Zero means none.
}

// Device access permission entry.
//
// Enforced via BPF_PROG_TYPE_CGROUP_DEVICE. Each entry whitelists access
// to a character or block device by major:minor number.
type cgroupDevice struct {
	Type   deviceKind `codec:"type,omitempty"`   // "c" (char), "b" (block), or "a" (all).
	Major  uint32     `codec:"major,omitempty"`  // Device major number.
	Minor  uint32     `codec:"minor,omitempty"`  // Device minor number.
	Access string     `codec:"access,omitempty"` // Combination of "r", "w", "m". Empty means no access.
}

// Pressure stall information (PSI) triggers.
//
// Monitors resource contention by firing notifications when stall time
// exceeds a threshold within a sliding window. Covers the cpu.pressure,
// memory.pressure, and io.pressure interface files.
type cgroupPSI struct {
	CPU    []cgroupPSITrigger `codec:"cpu,omitempty"`    // cpu.pressure triggers.
	Memory []cgroupPSITrigger `codec:"memory,omitempty"` // memory.pressure triggers.
	IO     []cgroupPSITrigger `codec:"io,omitempty"`     // io.pressure triggers.
}

// A single PSI pressure trigger.
//
// Fires when cumulative stall time within a monitoring window exceeds a
// configured threshold, for either partial or total task stalls.
type cgroupPSITrigger struct {
	Kind      psiKind `codec:"kind"`      // "some" or "full".
	Threshold uint64  `codec:"threshold"` // Stall threshold in microseconds.
	Window    uint64  `codec:"window"`    // Monitoring window in microseconds.
}
