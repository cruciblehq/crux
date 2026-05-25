package manifest

import "github.com/cruciblehq/crux/crex"

// Cloud network perimeter configuration.
//
// Expresses what flows are permitted at the cloud boundary. No flows are
// implied; every allowed path must be declared explicitly.
type Network struct {

	// Inbound flows allowed at the cloud perimeter.
	Ingress []IngressRule `codec:"ingress,omitempty"`

	// Outbound flows allowed at the cloud perimeter.
	Egress []EgressRule `codec:"egress,omitempty"`
}

// Validates all network rules.
func (n *Network) Validate() error {
	for i := range n.Ingress {
		if err := n.Ingress[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidProtocol, "ingress[%d]: %w", i, err)
		}
	}
	for i := range n.Egress {
		if err := n.Egress[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidProtocol, "egress[%d]: %w", i, err)
		}
	}
	return nil
}

// Inbound flow allowed at the cloud perimeter.
//
// Permits traffic from Source on Protocol/Port to reach the service.
type IngressRule struct {

	// Network protocol.
	//
	// One of "tcp" or "udp".
	Protocol string `codec:"protocol"`

	// Destination port on the service.
	//
	// Zero means any port.
	Port uint16 `codec:"port,omitempty"`

	// Traffic source.
	//
	// A CIDR (e.g. "0.0.0.0/0"), a Crucible service name, or "*" for any source.
	Source string `codec:"source,omitempty"`
}

// Validates that the rule specifies a recognised protocol.
func (r *IngressRule) Validate() error {
	return validateProtocol(r.Protocol)
}

// Outbound flow allowed at the cloud perimeter.
//
// Permits traffic from the service to Destination on Protocol/Port.
type EgressRule struct {

	// Network protocol.
	//
	// One of "tcp" or "udp".
	Protocol string `codec:"protocol"`

	// Destination port.
	//
	// Zero means any port.
	Port uint16 `codec:"port,omitempty"`

	// Traffic destination.
	//
	// A CIDR (e.g. "0.0.0.0/0"), a Crucible service name, or "*" for any destination.
	Destination string `codec:"destination,omitempty"`
}

// Validates that the rule specifies a recognised protocol.
func (r *EgressRule) Validate() error {
	return validateProtocol(r.Protocol)
}

// Checks that p is a recognised network protocol name.
func validateProtocol(p string) error {
	switch p {
	case "tcp", "udp":
		return nil
	default:
		return crex.Wrapf(ErrInvalidProtocol, "unknown protocol %q", p)
	}
}
