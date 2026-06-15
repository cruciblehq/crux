package net

import (
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Outbound connection produced by a .net egress grant.
//
// Names a host or Crucible service that this container must be able to reach.
// Destinations are matched literally with two exceptions: a bare "*" means
// unrestricted egress, and a leading "*." denotes all subdomains of a given
// domain (e.g. "*.crucible.com" matches any subdomain of crucible.com, such as
// api.crucible.com). No other glob patterns are accepted. This keeps the egress
// allowlist auditable while sparing services from enumerating every subdomain
// of a provider.
type EgressRule struct {

	// Network protocol.
	//
	// A protocol keyword, the "ip" wildcard for any protocol, or a decimal IP
	// protocol number. Keyword names and numbers follow the IANA Protocol
	// Numbers registry; see the net package documentation for the accepted
	// keyword set.
	Protocol string `codec:"protocol"`

	// Destination port.
	//
	// Meaningful only for the port-based protocols (tcp, udp, sctp, dccp),
	// where zero means any port and a non-zero value restricts the allow to
	// that specific port. Zero for all other protocols.
	Port uint16 `codec:"port,omitempty"`

	// Named destination: an FQDN, a subdomain wildcard, a Crucible service
	// name, a CIDR, or "*".
	//
	// External hosts are fully-qualified domains (e.g. "api.crucible.com"). A
	// leading "*." matches a domain and all subdomains. Internal services in
	// the same deployment are referenced by their short service name (e.g.
	// "postgres"). CIDR ranges (e.g. "10.0.0.0/8") address a block of hosts.
	// A bare "*" means no destination restriction.
	Destination string `codec:"destination"`
}

// Validates the egress rule.
//
// The protocol must be recognized, and a port is permitted only if the protocol
// supports one. The destination must be non-empty. A destination containing "*"
// must be either the bare "*" (unrestricted) or a single leading-label wildcard
// of the form "*.domain".
func (r *EgressRule) Validate() error {
	if !IsValidProtocol(r.Protocol) {
		return crex.Wrapf(ErrInvalidEgressRule, "unknown protocol %q", r.Protocol)
	}
	if !isPortBased(r.Protocol) && r.Port != 0 {
		return crex.Wrapf(ErrInvalidEgressRule, "protocol %q does not take a port", r.Protocol)
	}
	if r.Destination == "" {
		return crex.Wrapf(ErrInvalidEgressRule, "destination is empty")
	}
	if r.Destination != "*" && strings.Contains(r.Destination, "*") {
		suffix, ok := strings.CutPrefix(r.Destination, "*.")
		if !ok || suffix == "" || strings.Contains(suffix, "*") {
			return crex.Wrapf(ErrInvalidEgressRule, "wildcard destination %q must be \"*\" or of the form \"*.domain\"", r.Destination)
		}
	}
	return nil
}
