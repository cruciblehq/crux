// Package affordance compiles a service's security grants into a runtime spec.
//
// Each Crucible service is confined by a [Spec] compiled exclusively from the
// grants declared in its affordance. The spec encodes every permission the
// service holds across all kernel enforcement layers, and the executor applies
// it when starting the container.
//
// A grant has the form:
//
//	.SUBSYSTEM ARG... [KEY=VALUE...] [where CONDITION]
//
// The leading dot and name select the [subsystem.Subsystem] that handles the
// grant. Positional arguments carry the primary parameters, keyword arguments
// carry optional modifiers, and the optional where clause is a predicate
// evaluated against a hook's runtime arguments (supported only by .mac).
//
// The builder iterates grants in order and calls [subsystem.Subsystem.Build]
// to fold each grant into the spec. Repeated declarations are resolved by the
// target subsystem, which may treat them as a no-op, merge them, or reject
// them as a conflict depending on the subsystem's semantics.
//
// The following subsystems are currently implemented:
//
//	.cap        Linux capabilities
//	.fcap       file capabilities applied as xattrs on container binaries
//	.mac        BPF-LSM hook allow rules
//	.seccomp    syscall allow rules over the deny-all BPF filter
//	.rlimit     POSIX resource limits
//	.cgroup     cgroup v2 resource controls and device access
//	.mount      kernel pseudo-filesystem mounts
//	.device     device node provisioning
//	.net        nftables ingress and egress rules
//	.volume     persistent storage volumes
//	.kernel     host kernel requirements
//	.provision  compute resource envelope for scheduling
package affordance
