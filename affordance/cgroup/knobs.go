package cgroup

// Dotted knob paths used in cgroup v2 grants and OCI Unified entries.
//
// These mirror the names exposed by the cgroup v2 filesystem. Centralising
// them keeps spelling consistent across parsers, conflict messages, and the
// Subsystem-level dispatcher that routes knob writes to typed model fields.
const (
	subtreeControlKnob = "cgroup.subtree_control" // List of controllers delegated to child cgroups.
	devicesKnob        = "devices"                // Devices controller root knob (BPF allowlist).
	ioWeightKnob       = "io.weight"              // IO weight, scalar or per-device.
	ioMaxKnob          = "io.max"                 // IO bandwidth and IOPS caps per device.
	ioLatencyKnob      = "io.latency"             // IO latency target per device.
	ioCostModelKnob    = "io.cost.model"          // IO cost model per device.
	ioCostQoSKnob      = "io.cost.qos"            // IO cost quality-of-service per device.
	dmemMaxKnob        = "dmem.max"               // Device memory hard limit per region.
	dmemMinKnob        = "dmem.min"               // Device memory soft floor per region.
	dmemLowKnob        = "dmem.low"               // Device memory best-effort floor per region.
	psiCPUKnob         = "cpu.pressure"           // PSI triggers for CPU pressure.
	psiMemoryKnob      = "memory.pressure"        // PSI triggers for memory pressure.
	psiIOKnob          = "io.pressure"            // PSI triggers for IO pressure.
	hugeTLBKnob        = "hugetlb"                // Huge page limits per page size.
	rdmaKnob           = "rdma.max"               // RDMA resource limits per HCA device.
	miscKnob           = "misc.max"               // Miscellaneous scalar resource limits.
)
