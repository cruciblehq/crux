package manifest

import (
	"errors"
	"testing"
)

func TestPlanValidateOK(t *testing.T) {
	p := &Plan{
		Version:    PlanVersion,
		Services:   []Ref{{Target: "ns/x"}},
		Compute:    []Compute{{ID: "c1", Provider: "local"}},
		Containers: []Container{{Service: "s", Compute: "c1"}},
		Gateway:    Gateway{Routes: []Route{{Pattern: "/api", Service: "s"}}},
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
	p := &Plan{Version: PlanVersion, Services: []Ref{{}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlanValidatePropagatesComputeError(t *testing.T) {
	p := &Plan{Version: PlanVersion, Compute: []Compute{{Provider: "local"}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlanValidatePropagatesContainerError(t *testing.T) {
	p := &Plan{Version: PlanVersion, Containers: []Container{{Compute: "c1"}}}
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
