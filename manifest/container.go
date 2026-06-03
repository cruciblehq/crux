package manifest

import (
	"github.com/cruciblehq/crux/crex"
	affcap "github.com/cruciblehq/crux/security/fcap"
	affmac "github.com/cruciblehq/crux/security/mac"
	afnet "github.com/cruciblehq/crux/security/net"
	afvolume "github.com/cruciblehq/crux/security/volume"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

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
	Fcap affcap.Fcap `codec:"fcap"`

	// MAC (LSM) hook allow rules.
	//
	// Compiled from .mac grants and loaded as LSM policy into the container's
	// security module context. Each rule names a kernel hook and an optional
	// predicate that must match for the allow to apply. An empty Rules slice
	// means no LSM allows are granted.
	MAC affmac.MAC `codec:"mac"`

	// Container-level network policy.
	//
	// Compiled from .net grants and applied as nftables rules injected into the
	// container's network namespace by the VM init process before the container
	// starts. Rules enforce ingress ports and egress destinations within the
	// netns. Empty Ingress and Egress slices mean a deny-all baseline (no ports
	// are reachable inbound and no outbound destinations are reachable).
	Network afnet.NetworkPolicy `codec:"network"`

	// Persistent storage volumes declared for this container.
	//
	// Compiled from .volume grants and provisioned by the executor before the
	// container starts. Each entry maps to a directory or managed disk mounted
	// at Mount.Destination inside the container. An empty slice means no
	// persistent volumes are declared for this container.
	Volumes []afvolume.Mount `codec:"volumes,omitempty"`
}

// Validates the compiled container enforcement spec.
//
// Delegates to each sub-spec in order: Fcap, MAC, then Network.
func (c *Container) Validate() error {
	if err := c.Fcap.Validate(); err != nil {
		return crex.Wrap(ErrInvalidContainer, err)
	}
	if err := c.MAC.Validate(); err != nil {
		return crex.Wrap(ErrInvalidContainer, err)
	}
	if err := c.Network.Validate(); err != nil {
		return crex.Wrap(ErrInvalidContainer, err)
	}
	for i := range c.Volumes {
		if err := c.Volumes[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidContainer, err)
		}
	}
	return nil
}
