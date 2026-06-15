// Package net implements the container network subsystem.
//
// Net grants declare the network footprint of a container: which ports accept
// inbound connections and which destinations it is permitted to reach outbound.
// The blueprint builder reads the accumulated [Spec] and generates nftables
// rules that enforce a deny-all baseline modified only by the declared grants.
// Every outbound destination must be declared explicitly. Destinations can be
// specific FQDNs, subdomain wildcards (e.g. "*.crucible.com"), CIDR ranges,
// internal service names, or a bare "*" for unrestricted egress.
//
// The protocol is given as a keyword (tcp, udp, sctp, dccp, icmp, icmpv6, gre,
// esp, ah), the "ip" wildcard for any protocol, or the two-token form "proto N"
// for an arbitrary IP protocol number (0–255). Keyword names and numbers follow
// the IANA registry (https://www.iana.org/assignments/protocol-numbers); a
// number that identifies a recognized protocol is normalized to its keyword
// (e.g. "proto 6" becomes "tcp"). A port applies only to protocols that support
// them (tcp, udp, sctp, dccp); it is required for ingress and optional for
// egress, must be in the range 1–65535, and must be omitted for all other
// protocols. An omitted egress port allows any port on the destination,
// which is the broadest allowance the grant can express; against the deny-all
// baseline a specific port is preferred and the omitted form is reserved for
// protocols whose port is negotiated at runtime (e.g. FTP data channels,
// SIP/RTP media, RPC/portmapper services).
//
// Declare a port that will accept inbound connections:
//
//	.net ingress tcp 8080
//	.net ingress udp 53
//	.net ingress icmp
//	.net ingress proto 89
//
// Declare an outbound destination:
//
//	.net egress tcp 443 api.crucible.com
//	.net egress tcp 443 *.crucible.com
//	.net egress tcp 5432 postgres
//	.net egress tcp 443 "10.0.0.0/8"
//	.net egress tcp 443 *
//	.net egress icmp *
//	.net egress gre "10.0.0.0/8"
//	.net egress proto 89 mesh-router
package net
