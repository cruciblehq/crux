package blueprint

import (
	"testing"

	"github.com/cruciblehq/spec/affordance/net"
	"github.com/cruciblehq/spec/manifest"
)

func TestDeriveNetworks(t *testing.T) {
	p := &manifest.Plan{
		Infrastructure: manifest.Infrastructure{
			Computes: map[string]manifest.Compute{
				"c1": {},
				"c2": {},
			},
		},
		Containers: map[string]manifest.Container{
			"ctr1": {Network: net.Spec{
				Ingress: []net.IngressRule{{Protocol: "tcp", Port: 8080}},
				Egress:  []net.EgressRule{{Protocol: "tcp", Port: 443, Destination: "api.example.com"}},
			}},
		},
		Deployments: []manifest.Deployment{
			{Compute: "c1", Container: "ctr1"},
			{Compute: "c2", Container: "missing"},
		},
	}

	deriveNetworks(p)

	// Every compute referenced by a deployment gets a network entry.
	if len(p.Infrastructure.Networks) != 2 {
		t.Fatalf("derived networks = %d, want 2", len(p.Infrastructure.Networks))
	}

	// c1 inherits the container's ingress and egress rules.
	n1 := p.Infrastructure.Networks["c1"]
	if len(n1.Ingress) != 1 || n1.Ingress[0].Port != 8080 {
		t.Errorf("c1 ingress = %+v, want one rule on port 8080", n1.Ingress)
	}
	if len(n1.Egress) != 1 || n1.Egress[0].Destination != "api.example.com" {
		t.Errorf("c1 egress = %+v, want one rule to api.example.com", n1.Egress)
	}

	// c2 has no matching container and stays a deny-all (empty) baseline.
	n2 := p.Infrastructure.Networks["c2"]
	if len(n2.Ingress) != 0 || len(n2.Egress) != 0 {
		t.Errorf("c2 network = %+v, want empty deny-all baseline", n2)
	}

	// Each deployment is bound to its compute's network.
	for i := range p.Deployments {
		if p.Deployments[i].Network != p.Deployments[i].Compute {
			t.Errorf("deployment[%d] network = %q, want %q", i, p.Deployments[i].Network, p.Deployments[i].Compute)
		}
	}
}

func TestMergeIngress(t *testing.T) {
	var n manifest.Network
	rule := manifest.IngressRule{Protocol: "tcp", Port: 443, Source: "*"}

	mergeIngress(&n, rule)
	mergeIngress(&n, rule) // duplicate, must be ignored
	if len(n.Ingress) != 1 {
		t.Fatalf("after duplicate merge, len = %d, want 1", len(n.Ingress))
	}

	mergeIngress(&n, manifest.IngressRule{Protocol: "tcp", Port: 80, Source: "*"})
	if len(n.Ingress) != 2 {
		t.Fatalf("after distinct merge, len = %d, want 2", len(n.Ingress))
	}
}

func TestMergeEgress(t *testing.T) {
	var n manifest.Network
	rule := manifest.EgressRule{Protocol: "tcp", Port: 443, Destination: "api.example.com"}

	mergeEgress(&n, rule)
	mergeEgress(&n, rule) // duplicate, must be ignored
	if len(n.Egress) != 1 {
		t.Fatalf("after duplicate merge, len = %d, want 1", len(n.Egress))
	}

	mergeEgress(&n, manifest.EgressRule{Protocol: "tcp", Port: 443, Destination: "db.example.com"})
	if len(n.Egress) != 2 {
		t.Fatalf("after distinct merge, len = %d, want 2", len(n.Egress))
	}
}
