package net

import "github.com/cruciblehq/crux/crex"

// Container-level network spec compiled from .net grants.
//
// The platform enforces a deny-all baseline, and only the ports declared in
// Ingress are reachable inbound, and only the destinations declared in Egress
// are reachable outbound. Every destination must be declared explicitly using
// FQDNs, subdomain wildcards, CIDR ranges, internal service names, or a bare
// "*" for unrestricted egress.
type Spec struct {

	// Inbound ports.
	//
	// Only traffic arriving on a declared port and protocol is admitted into
	// the container's network namespace. All other inbound traffic is blocked.
	Ingress []IngressRule `codec:"ingress,omitempty"`

	// Outbound connections this container may make.
	//
	// Every entry names a destination: an FQDN, a subdomain wildcard, a CIDR
	// range, a Crucible service name, or a bare "*" for unrestricted egress.
	// The platform drops all other egress.
	Egress []EgressRule `codec:"egress,omitempty"`
}

// Validates the network spec.
//
// Each ingress and egress rule must be structurally valid.
func (s *Spec) Validate() error {
	for i := range s.Ingress {
		if err := s.Ingress[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidSpec, err)
		}
	}
	for i := range s.Egress {
		if err := s.Egress[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidSpec, err)
		}
	}
	return nil
}
