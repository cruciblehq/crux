package cgroup

import "reflect"

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
	PSI     psi       `knob:"psi" json:"psi,omitzero"`          // Pressure stall information triggers.
	HugeTLB []hugeTLB `knob:"hugetlb" json:"hugetlb,omitempty"` // Huge page limits per page size.
	RDMA    []rdma    `knob:"rdma" json:"rdma,omitempty"`       // RDMA resource limits per device.
	Misc    []misc    `knob:"misc" json:"misc,omitempty"`       // Miscellaneous scalar resource limits.
	Devices []device  `knob:"devices" json:"devices,omitempty"` // Device access permissions (BPF_PROG_TYPE_CGROUP_DEVICE).
	Dmem    []dmem    `knob:"dmem" json:"dmem,omitempty"`       // Device memory limits per region.

	// Tracks scalar knobs that have been explicitly set during a Build session.
	//
	// Used by the per-grant Build path to distinguish "still at default"
	// (zero value) from "explicitly set to value X" when detecting conflicts.
	// Excluded from codec serialisation; only the spec values are persisted.
	set map[string]struct{} `json:"-"`
}

// Returns a cgroup spec with restrictive defaults applied.
func newSpec() *spec {
	s := &spec{set: make(map[string]struct{})}
	setDefaults(reflect.TypeFor[spec](), reflect.ValueOf(s).Elem())
	return s
}
