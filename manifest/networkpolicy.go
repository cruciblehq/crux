package manifest

// Container-level network policy compiled from .net grants.
//
// The platform enforces a deny-all baseline relaxed by the declared ingress
// and egress rules. The builder compiles .net grants into this policy, which
// is then applied by nftables rules injected into the container's network
// namespace during VM initialisation. Every outbound destination must be
// declared explicitly — FQDNs, CIDR ranges, internal service names, or the
// keyword "any".

// Applied by nftables rules injected into each container's network namespace
// during VM initialisation. The platform enforces a deny-all baseline: only
// the ports declared in Listen are reachable inbound, and only the destinations
// declared in Egress are reachable outbound.
//
// Every outbound destination must be declared explicitly — FQDNs, CIDR ranges,
// internal service names, or the keyword "any". The network footprint is audited
// before marketplace approval.
type NetworkPolicy struct {

	// Inbound ports this container listens on.
	//
	// Only traffic arriving on a declared port and protocol is admitted into
	// the container's network namespace. Any inbound port not listed here is
	// blocked at the VM's nftables layer regardless of what the container
	// process actually binds.
	Listen []NetListenRule `codec:"listen,omitempty"`

	// Outbound connections this container may make.
	//
	// Every entry names a destination: an FQDN, a CIDR range, a Crucible
	// service name, or the keyword "any". The platform drops all other egress.
	Egress []NetEgressRule `codec:"egress,omitempty"`
}

// Inbound port declaration produced by a .net listen grant.
//
// Declares that this container accepts connections on the given protocol and
// port. The platform uses this to open exactly the listed ports — no more.
type NetListenRule struct {

	// Network protocol.
	//
	// One of "tcp" or "udp".
	Protocol string `codec:"protocol"`

	// Port number the container listens on.
	//
	// Must be a non-zero value in the range 1–65535.
	Port uint16 `codec:"port"`
}

// Outbound connection declaration produced by a .net egress grant.
//
// Names a specific host or Crucible service that this container must be able
// to reach. Wildcards are never permitted; every outbound destination is
// stated explicitly so that the platform can enforce a strict egress allowlist
// and marketplace users can audit the full network footprint before installing
// a service.
type NetEgressRule struct {

	// Network protocol.
	//
	// One of "tcp" or "udp".
	Protocol string `codec:"protocol"`

	// Destination port.
	//
	// Zero means any port to this destination is permitted. A non-zero value
	// restricts the allow to that specific port.
	Port uint16 `codec:"port,omitempty"`

	// Named destination: an FQDN, a Crucible service name, a CIDR, or "any".
	//
	// External hosts are fully-qualified domain names (e.g. "api.stripe.com").
	// Internal services in the same deployment are referenced by their short
	// service name (e.g. "postgres"). CIDR ranges (e.g. "10.0.0.0/8") address a
	// block of hosts. The keyword "any" means no destination restriction.
	//
	// Required. No implicit default.
	Destination string `codec:"destination"`
}
