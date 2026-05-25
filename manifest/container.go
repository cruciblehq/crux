package manifest

import specs "github.com/opencontainers/runtime-spec/specs-go"

// Compiled enforcement model for containers.
//
// This model is produced by the affordance builder during blueprint build and
// stored in the plan's Containers map. It covers the enforcement surface that
// runc applies at container creation time (OCI runtime spec, file capabilities,
// MAC hook allow rules, and container-level network policy. An empty value for
// any field is the explicit zero-grant baseline for that subsystem.
type Container struct {

	// Full OCI runtime spec passed to runc at container creation.
	//
	// Contains namespaces, capabilities, seccomp filter, cgroup device rules,
	// mounts, environment, and every other runc-level knob. This is Populated
	// and narrowed by the OCI subsystems (cap, seccomp, cgroup, rlimit, etc).
	OCI specs.Spec `codec:"oci"`

	// File capabilities.
	//
	// Compiled from .fcap grants and applied by the executor as xattrs on the
	// container filesystem before runc starts; runc then reads them during
	// execve to set the process capability sets. An empty Entries map means
	// no file capabilities are granted.
	Fcap FcapSpec `codec:"fcap"`

	// MAC (LSM) hook allow rules.
	//
	// Compiled from .mac grants and loaded as LSM policy into the container's
	// security module context. Each rule names a kernel hook and an optional
	// predicate that must match for the allow to apply. An empty Rules slice
	// means no LSM allows are granted.
	MAC MACSpec `codec:"mac"`

	// Container-level network policy.
	//
	// Compiled from .net grants and applied as nftables rules injected into the
	// container's network namespace by the VM init process before the container
	// starts. Rules enforce listen ports and egress destinations within the
	// netns. Empty Listen and Egress slices mean a deny-all baseline (no ports
	// are reachable inbound and no outbound destinations are reachable).
	Network NetworkPolicy `codec:"network"`
}
