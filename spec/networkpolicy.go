package spec

// Container-level network policy compiled from .net grants.
//
// The platform enforces a deny-all baseline: only the ports declared in
// Ingress are reachable inbound, and only the destinations declared in
// Egress are reachable outbound. Every outbound destination must be declared
// explicitly using FQDNs, CIDR ranges, internal service names, or "any".
type NetworkPolicy struct {

	// Inbound ports.
	//
	// Only traffic arriving on a declared port and protocol is admitted into
	// the container's network namespace. All other inbound traffic is blocked.
	Ingress []NetworkIngressRule `codec:"ingress,omitempty"`

	// Outbound connections this container may make.
	//
	// Every entry names a destination: an FQDN, a CIDR range, a Crucible
	// service name, or the keyword "any". The platform drops all other egress.
	Egress []NetworkEgressRule `codec:"egress,omitempty"`
}

// Inbound port declaration produced by a .net ingress grant.
//
// Declares that this container accepts connections on the given protocol and
// port. The platform opens exactly the listed ports.
type NetworkIngressRule struct {

	// Network protocol.
	//
	// One of "tcp" or "udp".
	Protocol string `codec:"protocol"`

	// Inbound port number.
	//
	// Must be in the range 1–65535.
	Port uint16 `codec:"port"`
}

// Outbound connection declaration produced by a .net egress grant.
//
// Names a specific host or Crucible service that this container must be able
// to reach. Wildcards are never permitted; every outbound destination is
// stated explicitly so that the platform can enforce a strict egress allowlist
// and marketplace users can audit the full network footprint before installing
// a service.
type NetworkEgressRule struct {

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
