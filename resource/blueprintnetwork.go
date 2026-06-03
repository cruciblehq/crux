package resource

import "github.com/cruciblehq/crux/manifest"

// Derives cloud-level perimeter network entries from per-container network
// policies declared via .net grants.
//
// Groups deployments by compute unit, unions the NetIngressRule and NetEgressRule
// declarations from each container's NetworkPolicy, and writes the aggregated
// result into Infrastructure.Networks keyed by compute ID. Computes with no
// container network declarations receive an empty Network entry, which produces
// a deny-all cloud perimeter baseline. Each Deployment.Network is set to match
// Deployment.Compute so that plan validation resolves correctly.
func deriveNetworks(p *manifest.Plan) {
	nets := make(map[string]*manifest.Network, len(p.Infrastructure.Computes))

	// Initialise an entry for every compute referenced by a deployment.
	// This ensures that computes with no network grants still get a cloud
	// perimeter entry, making the deny-all baseline explicit for plan
	// validation.
	for i := range p.Deployments {
		id := p.Deployments[i].Compute
		if _, ok := nets[id]; !ok {
			nets[id] = &manifest.Network{}
		}
	}

	// Accumulate each container's network policy into its compute's network.
	for i := range p.Deployments {
		dep := &p.Deployments[i]
		ctr, ok := p.Containers[dep.Container]
		if !ok {
			continue
		}
		net := nets[dep.Compute]
		policy := ctr.Network
			for _, r := range policy.Ingress {
			mergeIngress(net, manifest.IngressRule{
				Protocol: r.Protocol,
				Port:     r.Port,
			})
		}
		for _, r := range policy.Egress {
			mergeEgress(net, manifest.EgressRule{
				Protocol:    r.Protocol,
				Port:        r.Port,
				Destination: cloudDest(r.Destination),
			})
		}
	}

	// Write derived network entries into the plan and bind deployments.
	if p.Infrastructure.Networks == nil {
		p.Infrastructure.Networks = make(map[string]manifest.Network, len(nets))
	}
	for id, net := range nets {
		p.Infrastructure.Networks[id] = *net
	}
	for i := range p.Deployments {
		p.Deployments[i].Network = p.Deployments[i].Compute
	}
}

// Maps the container-level "any" destination keyword to the cloud-level "*".
//
// Container grant syntax uses "any" to declare unrestricted egress. Cloud
// perimeter rules use "*" for the same meaning. All other values pass through
// unchanged.
func cloudDest(d string) string {
	if d == "any" {
		return "*"
	}
	return d
}

// Appends rule to net.Ingress unless an identical rule is already present.
func mergeIngress(net *manifest.Network, rule manifest.IngressRule) {
	for _, r := range net.Ingress {
		if r.Protocol == rule.Protocol && r.Port == rule.Port && r.Source == rule.Source {
			return
		}
	}
	net.Ingress = append(net.Ingress, rule)
}

// Appends rule to net.Egress unless an identical rule is already present.
func mergeEgress(net *manifest.Network, rule manifest.EgressRule) {
	for _, r := range net.Egress {
		if r.Protocol == rule.Protocol && r.Port == rule.Port && r.Destination == rule.Destination {
			return
		}
	}
	net.Egress = append(net.Egress, rule)
}
