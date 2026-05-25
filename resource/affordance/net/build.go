package net

import (
	"fmt"
	"strconv"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/spec"
	"github.com/cruciblehq/crux/resource/affordance/agl"
	"github.com/cruciblehq/crux/resource/affordance/subsystem"
)

// Operation tokens accepted as the first positional argument.
const (
	opIngress = "ingress" // Declare an inbound port.
	opEgress = "egress" // Declare an explicit outbound destination.
)

// Protocol tokens accepted as the second positional argument.
const (
	protoTCP = "tcp"
	protoUDP = "udp"
)

// Implementation of the container network policy subsystem.
//
// Holds a pointer to the accumulated [Spec] wired in at construction time.
// Each Build call appends one ingress or egress declaration; declarations are
// additive. The deduplication key prevents the same declaration from appearing
// more than once.
type Subsystem struct {
	spec *Spec
}

// Returns a Subsystem wired to accumulate into s.
func New(s *Spec) *Subsystem {
	return &Subsystem{spec: s}
}

// Returns the net subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameNet
}

// Returns the deduplication key for a net grant.
//
// For ingress grants the key encodes the operation, protocol, and port. For
// egress grants it also encodes the destination. Identical declarations are
// rejected as conflicts by the builder before Build is called.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) < 2 {
		return ""
	}
	op := g.Args[0].Value
	switch op {
	case opIngress:
		if len(g.Args) < 3 {
			return ""
		}
		return fmt.Sprintf("ingress:%s:%s", g.Args[1].Value, g.Args[2].Value)
	case opEgress:
		if len(g.Args) < 4 {
			return ""
		}
		return fmt.Sprintf("egress:%s:%s:%s", g.Args[1].Value, g.Args[2].Value, g.Args[3].Value)
	default:
		return ""
	}
}

// Applies a parsed grant to the accumulated spec.
//
// Two operations are supported.
//
// An ingress grant has the form ".net ingress PROTOCOL PORT". PROTOCOL is "tcp"
// or "udp". PORT is a non-zero port number in the range 1–65535. It declares
// that this container accepts inbound connections on that port; the platform
// blocks all other inbound traffic.
//
// An egress grant has the form ".net egress PROTOCOL PORT DESTINATION".
// DESTINATION is a required FQDN (e.g. "api.stripe.com") or the short name of
// another Crucible service in the same deployment (e.g. "postgres"). Wildcards
// are never accepted. The platform generates a deny-all egress rule and then
// adds an allow entry for each declared destination, making the full set of
// outbound connections auditable.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	op := g.Args[0].Value
	proto := g.Args[1].Value
	port, err := parsePort(g.Args[2], op)
	if err != nil {
		return err
	}
	switch op {
	case opIngress:
		s.spec.Ingress = append(s.spec.Ingress, spec.NetworkIngressRule{
			Protocol: proto,
			Port:     port,
		})
	case opEgress:
		s.spec.Egress = append(s.spec.Egress, spec.NetworkEgressRule{
			Protocol:    proto,
			Port:        port,
			Destination: g.Args[3].Value,
		})
	}
	return nil
}

// Validates the structural shape of a net grant.
//
// Both operations share the no-where-clause and no-kwargs constraints. Listen
// takes exactly three positional arguments (operation, protocol, port). Egress
// takes exactly four (operation, protocol, port, destination). The operation,
// protocol, and destination argument types are also checked here.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in net grant")
	}
	if len(g.Kwargs) > 0 {
		return crex.Wrapf(ErrInvalidGrant, "unexpected keyword arguments in net grant")
	}
	if len(g.Args) < 1 || g.Args[0].Type != agl.ArgName {
		return crex.Wrapf(ErrInvalidGrant, "first argument must be an operation (ingress, egress)")
	}
	op := g.Args[0].Value
	switch op {
	case opIngress:
		if len(g.Args) != 3 {
			return crex.Wrapf(ErrInvalidGrant, "ingress grant requires exactly three arguments (ingress protocol port)")
		}
	case opEgress:
		if len(g.Args) != 4 {
			return crex.Wrapf(ErrInvalidGrant, "egress grant requires exactly four arguments (egress protocol port destination)")
		}
	default:
		return crex.Wrapf(ErrInvalidGrant, "unknown operation %q in net grant", op)
	}
	if g.Args[1].Type != agl.ArgName {
		return crex.Wrapf(ErrInvalidGrant, "second argument must be a protocol (tcp, udp)")
	}
	switch g.Args[1].Value {
	case protoTCP, protoUDP:
	default:
		return crex.Wrapf(ErrInvalidGrant, "unknown protocol %q in net grant", g.Args[1].Value)
	}
	if op == opEgress {
		switch g.Args[3].Type {
		case agl.ArgName:
			// ArgName covers plain service names (e.g. "postgres"), dotted FQDNs
			// (e.g. "api.stripe.com"), and the explicit wildcard keyword "any".
		case agl.ArgStrASCII:
			// ArgStrASCII covers quoted CIDR strings (e.g. "10.0.0.0/8",
			// "0.0.0.0/0"). CIDRs must be quoted because the lexer cannot
			// tokenise them as bare words.
		default:
			return crex.Wrapf(ErrInvalidGrant, "destination must be a hostname, service name, CIDR, or \"any\" in net egress grant")
		}
	}
	return nil
}

// Parses a port argument into an unsigned 16-bit port number.
//
// For ingress grants the port must be non-zero (1–65535). For egress grants
// port 0 is accepted and means "any port to this destination".
func parsePort(a agl.Arg, op string) (uint16, error) {
	if a.Type != agl.ArgInt {
		return 0, crex.Wrapf(ErrInvalidGrant, "port must be an integer in net grant")
	}
	n, err := strconv.ParseUint(a.Value, 0, 16)
	if err != nil {
		return 0, crex.Wrapf(ErrInvalidGrant, "port out of range in net grant")
	}
	if op == opIngress && n == 0 {
		return 0, crex.Wrapf(ErrInvalidGrant, "ingress port must be non-zero")
	}
	return uint16(n), nil
}

