package net

import (
	"fmt"
	"strconv"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
	"github.com/cruciblehq/crux/crex"
)

// Operation tokens accepted as the first positional argument.
const (
	opIngress = "ingress" // Declare an inbound port.
	opEgress  = "egress"  // Declare an explicit outbound destination.
)

// Implementation of the container network subsystem.
//
// Holds a pointer to the accumulated [Spec]. Each Build call appends
// one ingress or egress declaration. The deduplication key prevents the same
// declaration from appearing more than once.
type Subsystem struct {
	spec *Spec // Write target for accumulated grants.
}

// Returns a Subsystem wired to accumulate into spec.
func New(spec *Spec) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the net subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameNet
}

// Returns the deduplication key for a net grant.
//
// For ingress grants the key encodes the operation, protocol, and port. For
// egress grants it also encodes the destination. The key is derived from the
// parsed grant so that equivalent declarations collapse regardless of how the
// optional port was written. Malformed grants yield an empty key and are
// reported by Build.
func (s *Subsystem) Key(g *agl.Model) string {
	p, err := parseGrant(g)
	if err != nil {
		return ""
	}
	if p.op == opIngress {
		return fmt.Sprintf("ingress:%s:%d", p.proto, p.port)
	}
	return fmt.Sprintf("egress:%s:%d:%s", p.proto, p.port, p.dest)
}

// Applies a parsed grant to the accumulated spec.
//
// An ingress grant has the form ".net ingress PROTOCOL [PORT]". PROTOCOL is a
// supported protocol keyword, the "ip" wildcard for any protocol, or the form
// "proto N" for an arbitrary IP protocol number; a number that identifies a
// recognized protocol is normalized to its keyword (e.g. "proto 6" becomes
// "tcp"). PORT is required and in the range 1–65535 for the port-based
// protocols (isPortBased) and must be omitted for all others. An egress grant
// has the form ".net egress PROTOCOL [PORT] DEST". PORT is optional for the
// port-based protocols (omitted means any port) and must be omitted for all
// others. Omitting the port is the broadest allowance an egress grant can
// express, so a specific port is preferred (omitted ports should be reserved
// for protocols whose port is negotiated at runtime and cannot be pinned, such
// as FTP data channels). DEST is a required FQDN (e.g.
// "api.crucible.com"), a subdomain wildcard (e.g. "*.crucible.com"), the name
// of another Crucible service in the same deployment (e.g. "postgres"), a
// quoted CIDR, or a bare "*" for unrestricted egress. Apart from "*" and the
// "*." prefix form, glob patterns are not accepted.
func (s *Subsystem) Build(g *agl.Model) error {
	p, err := parseGrant(g)
	if err != nil {
		return err
	}
	switch p.op {
	case opIngress:
		s.spec.Ingress = append(s.spec.Ingress, IngressRule{
			Protocol: p.proto,
			Port:     p.port,
		})
	case opEgress:
		s.spec.Egress = append(s.spec.Egress, EgressRule{
			Protocol:    p.proto,
			Port:        p.port,
			Destination: p.dest,
		})
	}
	return nil
}

// Parsed positional arguments of a net grant.
//
// Proto holds the normalized protocol value stored on the rule: a keyword, the
// "ip" wildcard, or a decimal IP protocol number. Port is zero when the
// protocol carries no port or when an egress port was omitted. Dest is empty
// for ingress grants.
type parsedGrant struct {
	op    string // opIngress or opEgress.
	proto string // Normalized protocol value.
	port  uint16 // Port number, or zero when none applies.
	dest  string // Egress destination.
}

// Parses and validates the positional arguments of a net grant.
//
// Both operations reject where clauses and keyword arguments. The protocol is
// either a single keyword token or the two-token "proto N" form. The port is
// accepted only for port-based protocols: required and non-zero for ingress,
// optional for egress. The egress destination argument type is also checked.
func parseGrant(g *agl.Model) (parsedGrant, error) {
	if g.Where != nil {
		return parsedGrant{}, crex.Wrapf(ErrInvalidGrant, "unexpected where clause in net grant")
	}
	if len(g.Kwargs) > 0 {
		return parsedGrant{}, crex.Wrapf(ErrInvalidGrant, "unexpected keyword arguments in net grant")
	}
	if len(g.Args) < 1 || g.Args[0].Type != agl.ArgName {
		return parsedGrant{}, crex.Wrapf(ErrInvalidGrant, "first argument must be an operation (ingress, egress)")
	}
	op := g.Args[0].Value
	if op != opIngress && op != opEgress {
		return parsedGrant{}, crex.Wrapf(ErrInvalidGrant, "unknown operation %q in net grant", op)
	}
	proto, rest, err := parseProtocol(g.Args[1:])
	if err != nil {
		return parsedGrant{}, err
	}
	p := parsedGrant{op: op, proto: proto}
	if op == opIngress {
		return p, parseIngressRest(&p, rest)
	}
	return p, parseEgressRest(&p, rest)
}

// Parses the protocol from the leading positional arguments.
//
// Returns the normalized protocol value and the remaining arguments after the
// protocol. The protocol is either a single keyword token or the two-token form
// "proto N" where N is an IP protocol number in the range 0–255. A number that
// identifies a recognized protocol is normalized to its keyword.
func parseProtocol(args []agl.Arg) (string, []agl.Arg, error) {
	if len(args) < 1 || args[0].Type != agl.ArgName {
		return "", nil, crex.Wrapf(ErrInvalidGrant, "second argument must be a protocol (tcp, udp, sctp, dccp, icmp, icmpv6, gre, esp, ah, ip, or \"proto N\")")
	}
	if args[0].Value == protoNumber {
		if len(args) < 2 || args[1].Type != agl.ArgInt {
			return "", nil, crex.Wrapf(ErrInvalidGrant, "proto requires a numeric protocol number")
		}
		n, err := strconv.ParseUint(args[1].Value, 0, 8)
		if err != nil {
			return "", nil, crex.Wrapf(ErrInvalidGrant, "protocol number out of range (0-255) in net grant")
		}
		return normalizeProtocolNumber(n), args[2:], nil
	}
	proto := args[0].Value
	if proto == protoAll {
		proto = protoIP
	}
	if !isPortBased(proto) && !isPortless(proto) && proto != protoIP {
		return "", nil, crex.Wrapf(ErrInvalidGrant, "unknown protocol %q in net grant", args[0].Value)
	}
	return proto, args[1:], nil
}

// Parses the arguments following the protocol for an ingress grant.
//
// Port-based protocols require exactly one trailing port argument in the
// range 1–65535. All other protocols take no trailing arguments.
func parseIngressRest(p *parsedGrant, rest []agl.Arg) error {
	if !isPortBased(p.proto) {
		if len(rest) != 0 {
			return crex.Wrapf(ErrInvalidGrant, "protocol %q takes no port in net ingress grant", p.proto)
		}
		return nil
	}
	if len(rest) != 1 {
		return crex.Wrapf(ErrInvalidGrant, "ingress grant for %q requires a port", p.proto)
	}
	port, err := parsePort(rest[0])
	if err != nil {
		return err
	}
	p.port = port
	return nil
}

// Parses the arguments following the protocol for an egress grant.
//
// Port-based protocols accept an optional port (omitted means any port)
// followed by a required destination. All other protocols take exactly
// one destination argument.
func parseEgressRest(p *parsedGrant, rest []agl.Arg) error {
	if !isPortBased(p.proto) {
		if len(rest) != 1 {
			return crex.Wrapf(ErrInvalidGrant, "egress grant for %q requires exactly one destination", p.proto)
		}
		return parseDestination(p, rest[0])
	}
	switch len(rest) {
	case 1:
		return parseDestination(p, rest[0])
	case 2:
		port, err := parsePort(rest[0])
		if err != nil {
			return err
		}
		p.port = port
		return parseDestination(p, rest[1])
	default:
		return crex.Wrapf(ErrInvalidGrant, "egress grant for %q accepts at most a port and a destination", p.proto)
	}
}

// Records the egress destination after checking its argument type.
//
// A destination is a service name or FQDN (ArgName) or the unrestricted "*",
// a subdomain wildcard, or a quoted CIDR (ArgStrASCII).
func parseDestination(p *parsedGrant, a agl.Arg) error {
	switch a.Type {
	case agl.ArgName, agl.ArgStrASCII:
		p.dest = a.Value
		return nil
	default:
		return crex.Wrapf(ErrInvalidGrant, "destination must be a hostname, service name, CIDR, or \"*\" in net egress grant")
	}
}

// Parses a port argument into an unsigned 16-bit port number.
//
// A port is in the range 1–65535; zero is not a valid port and is rejected.
func parsePort(a agl.Arg) (uint16, error) {
	if a.Type != agl.ArgInt {
		return 0, crex.Wrapf(ErrInvalidGrant, "port must be an integer in net grant")
	}
	n, err := strconv.ParseUint(a.Value, 0, 16)
	if err != nil {
		return 0, crex.Wrapf(ErrInvalidGrant, "port out of range in net grant")
	}
	if n == 0 {
		return 0, crex.Wrapf(ErrInvalidGrant, "port must be in the range 1-65535 in net grant")
	}
	return uint16(n), nil
}
