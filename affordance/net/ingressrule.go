package net

import "github.com/cruciblehq/crux/crex"

// Inbound port declaration produced by a .net ingress grant.
//
// Declares that this container accepts connections on the given protocol and
// port. The platform opens exactly the listed ports.
type IngressRule struct {

	// Network protocol.
	//
	// A protocol keyword, the "ip" wildcard for any protocol, or a decimal IP
	// protocol number. Keyword names and numbers follow the IANA Protocol
	// Numbers registry; see the net package documentation for the accepted
	// keyword set.
	Protocol string `codec:"protocol"`

	// Inbound port number.
	//
	// Non-zero (1–65535) for protocols supporting ports (tcp, udp, sctp,
	// dccp); zero for all other protocols.
	Port uint16 `codec:"port,omitempty"`
}

// Validates the ingress rule.
//
// The protocol must be recognized. Protocols that support ports require a non-
// zero port; all other protocols must leave the port as zero.
func (r *IngressRule) Validate() error {
	if !IsValidProtocol(r.Protocol) {
		return crex.Wrapf(ErrInvalidIngressRule, "unknown protocol %q", r.Protocol)
	}
	if isPortBased(r.Protocol) {
		if r.Port == 0 {
			return crex.Wrapf(ErrInvalidIngressRule, "port must be non-zero for protocol %q", r.Protocol)
		}
		return nil
	}
	if r.Port != 0 {
		return crex.Wrapf(ErrInvalidIngressRule, "protocol %q does not take a port", r.Protocol)
	}
	return nil
}
