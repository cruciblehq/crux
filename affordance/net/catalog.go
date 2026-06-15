package net

import "strconv"

// Protocols accepted.
const (
	protoTCP    = "tcp"    // TCP protocol (port-based).
	protoUDP    = "udp"    // UDP protocol (port-based).
	protoSCTP   = "sctp"   // SCTP protocol (port-based).
	protoDCCP   = "dccp"   // DCCP protocol (port-based).
	protoICMP   = "icmp"   // ICMP protocol (portless).
	protoICMPv6 = "icmpv6" // ICMPv6 protocol (portless).
	protoGRE    = "gre"    // GRE tunnelling protocol (portless).
	protoESP    = "esp"    // IPsec ESP protocol (portless).
	protoAH     = "ah"     // IPsec AH protocol (portless).
	protoIP     = "ip"     // Any IP protocol (wildcard, portless).
	protoAll    = "all"    // Alias for protoIP, normalized to "ip".
	protoNumber = "proto"  // Prefix introducing a numeric IP protocol number.
)

// Protocol keywords that carry a port number.
var portBasedProtocols = map[string]bool{
	protoTCP:  true,
	protoUDP:  true,
	protoSCTP: true,
	protoDCCP: true,
}

// Protocol keywords that do not carry a port number.
var portlessProtocols = map[string]bool{
	protoICMP:   true,
	protoICMPv6: true,
	protoGRE:    true,
	protoESP:    true,
	protoAH:     true,
}

// IANA protocol numbers for the recognized protocol keywords, used to normalize
// the "proto N" form to its keyword when N identifies a known protocol.
var protocolNumberToKeyword = map[uint64]string{
	1:   protoICMP,
	6:   protoTCP,
	17:  protoUDP,
	33:  protoDCCP,
	47:  protoGRE,
	50:  protoESP,
	51:  protoAH,
	58:  protoICMPv6,
	132: protoSCTP,
}

// Whether p is a protocol keyword that carries a port number.
func isPortBased(p string) bool {
	return portBasedProtocols[p]
}

// Whether p is a protocol keyword that does not carry a port number.
func isPortless(p string) bool {
	return portlessProtocols[p]
}

// Whether p is a stored protocol value: a recognized keyword, an IP protocol
// number in the range 0–255, or the "ip" wildcard.
func IsValidProtocol(p string) bool {
	if isPortBased(p) || isPortless(p) || p == protoIP {
		return true
	}
	_, err := strconv.ParseUint(p, 10, 8)
	return err == nil
}

// Normalizes an IANA protocol number to its protocol keyword when the number
// identifies a recognized protocol, otherwise returns its decimal form.
func normalizeProtocolNumber(n uint64) string {
	if kw, ok := protocolNumberToKeyword[n]; ok {
		return kw
	}
	return strconv.FormatUint(n, 10)
}
