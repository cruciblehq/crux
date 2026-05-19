package cgroup

// Process and thread count limit for the cgroup.
//
// The pids controller caps the number of tasks (processes plus threads) that
// can exist within the cgroup and its descendants combined.
type pids struct {
	Max uint64 `knob:"max" json:"max,omitempty"` // Maximum number of processes and threads allowed in the cgroup (0 prevents new tasks entirely).
}
