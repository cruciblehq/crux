// Package net implements the container network policy subsystem.
//
// Net grants declare the complete network footprint of a container: which ports
// it accepts inbound connections on, and exactly which destinations it is
// permitted to reach outbound. The blueprint builder reads the accumulated
// [Spec] and generates nftables rules that enforce a deny-all baseline modified
// only by the declared grants.
//
// This is the enforcement mechanism for the Crucible marketplace security model.
// Every outbound destination must be declared explicitly; the platform exposes
// that list to auditors before a service is approved. Destinations can be
// specific FQDNs, CIDR ranges, internal service names, or the keyword "any",
// but nothing is implicit: an "any" that the developer cannot justify, or a
// CIDR that auditors cannot account for, results in the service being rejected
// from the marketplace.
//
// Two operations are supported.
//
// An ingress grant declares a port the container will accept inbound
// connections on. The platform admits traffic only on the listed ports.
//
//	.net ingress tcp 8080
//	.net ingress tcp 443
//
// An egress grant declares an outbound destination. The destination is required
// and must be explicitly stated. Accepted forms: an FQDN, a CIDR (quoted), a
// Crucible service name, or the keyword "any".
//
//	.net egress tcp 443 api.stripe.com
//	.net egress tcp 5432 postgres
//	.net egress tcp 443 "10.0.0.0/8"
//	.net egress tcp 443 any
package net
