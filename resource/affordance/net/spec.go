package net

import "github.com/cruciblehq/crux/spec"

// Accumulated container network policy from .net grants.
//
// Collected by the subsystem during an affordance build session. The blueprint
// builder converts this to a [spec.NetworkPolicy] for inclusion in the
// compiled service spec.
type Spec struct {

	// Accumulated inbound port declarations.
	Ingress []spec.NetworkIngressRule

	// Accumulated outbound connection declarations.
	Egress []spec.NetworkEgressRule
}
