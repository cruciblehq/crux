package blueprint

import "github.com/cruciblehq/spec/manifest"

// Derives the network entries of the infrastructure from the network specs of
// the containers.
//
// Deployments are grouped by compute entry, and the network specs of all
// containers assigned to a compute are merged into a single entry written
// to Infrastructure.Networks keyed by compute ID. Ingress and egress rules
// are merged without duplicates. Computes that have no container network
// declarations still receive an empty entry, which denies all traffic at the
// perimeter. Each deployment's Network field is set to its Compute so plan
// validation resolves the reference correctly.
func deriveNetworks(p *manifest.Plan) {
	nets := make(map[string]*manifest.Network, len(p.Infrastructure.Computes))

	// Initialise an entry for every compute referenced by a deployment, so
	// computes without container network declarations still get an entry
	// that denies all traffic, which plan validation can see.
	for i := range p.Deployments {
		id := p.Deployments[i].Compute
		if _, ok := nets[id]; !ok {
			nets[id] = &manifest.Network{}
		}
	}

	// Accumulate each container's network spec into its compute's network.
	for i := range p.Deployments {
		dep := &p.Deployments[i]
		ctr, ok := p.Containers[dep.Container]
		if !ok {
			continue
		}
		net := nets[dep.Compute]
		spec := ctr.Network
		for _, r := range spec.Ingress {
			mergeIngress(net, manifest.IngressRule{
				Protocol: r.Protocol,
				Port:     r.Port,
			})
		}
		for _, r := range spec.Egress {
			mergeEgress(net, manifest.EgressRule{
				Protocol:    r.Protocol,
				Port:        r.Port,
				Destination: r.Destination,
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
