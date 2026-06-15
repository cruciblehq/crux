package manifest

import (
	"errors"
	"testing"

	afnet "github.com/cruciblehq/crux/affordance/net"
)

func TestPlanValidateOK(t *testing.T) {
	p := &Plan{
		Version:  PlanVersion,
		Services: map[string]string{"svc": "ns/x"},
		Infrastructure: Infrastructure{
			Computes: map[string]Compute{"c1": {Type: "local", Config: &ComputeLocal{Host: "localhost"}}},
			Networks: map[string]Network{"n1": {}},
		},
		Containers:  map[string]Container{"svc": {}},
		Deployments: []Deployment{{Service: "svc", Container: "svc", Compute: "c1", Network: "n1"}},
		Gateway:     Gateway{Routes: []Route{{Pattern: "/api", Service: "svc"}}},
	}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPlanValidateBadVersion(t *testing.T) {
	err := (&Plan{Version: 99}).Validate()
	if !errors.Is(err, ErrUnsupportedPlanVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedPlanVersion", err)
	}
}

func TestPlanValidatePropagatesServiceError(t *testing.T) {
	p := &Plan{Version: PlanVersion, Services: map[string]string{"svc": ""}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlanValidatePropagatesInfraError(t *testing.T) {
	p := &Plan{
		Version: PlanVersion,
		Infrastructure: Infrastructure{
			Networks: map[string]Network{
				"n1": {Ingress: []IngressRule{{Protocol: "bogus"}}},
			},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlanValidatePropagatesAssociationError(t *testing.T) {
	p := &Plan{Version: PlanVersion, Deployments: []Deployment{{Container: "c1", Compute: "c1", Network: "n1"}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlanValidatePropagatesGatewayError(t *testing.T) {
	p := &Plan{Version: PlanVersion, Gateway: Gateway{Routes: []Route{{}}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlanValidatePropagatesEnvironmentError(t *testing.T) {
	p := &Plan{
		Version:      PlanVersion,
		Environments: map[string]Environment{"e": {ID: "Bad Name"}},
	}
	err := p.Validate()
	if !errors.Is(err, ErrInvalidEnvironmentID) {
		t.Fatalf("err = %v, want ErrInvalidEnvironmentID", err)
	}
}

func TestPlanValidatePropagatesContainerError(t *testing.T) {
	p := &Plan{
		Version: PlanVersion,
		Containers: map[string]Container{
			"c1": {
				Network: afnet.Spec{
					Ingress: []afnet.IngressRule{{Protocol: "invalid", Port: 80}},
				},
			},
		},
	}
	err := p.Validate()
	if !errors.Is(err, ErrInvalidPlan) {
		t.Fatalf("err = %v, want ErrInvalidPlan", err)
	}
}
