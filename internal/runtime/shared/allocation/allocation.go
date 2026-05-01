package allocation

// Provisioned resources available to the spec.
//
// Allocation is the typed, read-only output of the provisioning subsystems.
// The cgroup spec validator consumes it to check that any identity-keyed
// reference in the spec (e.g. cpuset.cpus index 2) resolves to a resource
// that was actually provisioned. Allocation values use a local reference
// frame (logical indices into the provisioned set), not the kernel
// reference frame. Translation to kernel-frame identifiers happens at
// compose time.
type Allocation struct{}

// Returns the number of CPUs available in the allocation.
//
// Used by the validator to bound-check CPU indices in cpuset references.
func (a *Allocation) CPUCount() uint32 {
	return 0
}

// Returns the number of NUMA memory nodes available in the allocation.
//
// Used by the validator to bound-check memory-node indices in cpuset
// references.
func (a *Allocation) MemoryNodeCount() uint32 {
	return 0
}
