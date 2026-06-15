package cgroup

// Accumulated cgroup v2 resource spec.
//
// Controls CPU, memory, I/O, process count, CPU affinity, huge pages, RDMA,
// miscellaneous scalar resources, pressure stall thresholds, and hierarchy
// structure. Each controller is represented with typed fields and list
// entries that can be merged into another spec.
type spec struct {
	Cgroup  cgroup    `knob:"cgroup" json:"cgroup,omitzero"`    // cgroup hierarchy and freeze controls.
	CPU     cpu       `knob:"cpu" json:"cpu,omitzero"`          // CPU time and scheduling weight.
	Memory  memory    `knob:"memory" json:"memory,omitzero"`    // Memory limits and protections.
	IO      io        `knob:"io" json:"io,omitzero"`            // Block I/O weight, per-device limits, and cost model.
	PIDs    pids      `knob:"pids" json:"pids,omitzero"`        // Process count limit.
	CPUSet  cpuSet    `knob:"cpuset" json:"cpuset,omitzero"`    // CPU and memory node affinity.
	HugeTLB []hugeTLB `knob:"hugetlb" json:"hugetlb,omitempty"` // Huge page limits per page size.
	RDMA    []rdma    `knob:"rdma.max" json:"rdma,omitempty"`   // RDMA resource limits per device.
	Misc    []misc    `knob:"misc.max" json:"misc,omitempty"`   // Miscellaneous scalar resource limits.
	Devices []device  `knob:"devices" json:"devices,omitempty"` // Device access permissions (BPF_PROG_TYPE_CGROUP_DEVICE).
	Dmem    []dmem    `knob:"dmem" json:"dmem,omitempty"`       // Device memory limits per region.

	// Pressure stall information triggers.
	//
	// PSI has no "knob" tag because its kernel file names (cpu.pressure,
	// memory.pressure, io.pressure) share prefixes with the CPU, memory,
	// and IO controllers. Routing is handled by an explicit intercept in
	// [applyKnob] before the struct-tag walker runs.
	PSI psi `json:"psi,omitzero"`

	// Tracks scalar knobs that have been explicitly set during a Build session.
	//
	// Used by the per-grant Build path to detect conflicts when the same knob
	// is declared more than once. Excluded from codec serialisation; only the
	// spec values are persisted.
	seen map[string]bool `json:"-"`
}

// Returns a new cgroup spec.
func newSpec() *spec {
	return &spec{seen: make(map[string]bool)}
}
