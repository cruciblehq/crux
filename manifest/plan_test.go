package manifest

import (
	"errors"
	"testing"
)

func TestPlanValidateOK(t *testing.T) {
	p := &Plan{
		Version:     PlanVersion,
		Services:    map[string]string{"svc": "ns/x"},
		Compute:     map[string]Compute{"c1": {Provider: "local"}},
		Containers:  map[string]Container{"svc": {}},
		Deployments: []Deployment{{Service: "svc", Container: "svc", Compute: "c1"}},
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

func TestPlanValidatePropagatesComputeError(t *testing.T) {
	p := &Plan{Version: PlanVersion, Compute: map[string]Compute{"c1": {}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlanValidatePropagatesAssociationError(t *testing.T) {
	p := &Plan{Version: PlanVersion, Deployments: []Deployment{{Container: "c1", Compute: "c1"}}}
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
