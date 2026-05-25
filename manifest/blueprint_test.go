package manifest

import (
	"errors"
	"testing"
)

func TestBlueprintValidateOK(t *testing.T) {
	b := &Blueprint{
		Services: []Ref{
			{ID: "api", Ref: "ns/api"},
			{ID: "auth", Ref: "ns/auth"},
		},
		Gateway: Gateway{Routes: []Route{{Pattern: "/api", Service: "api"}}},
		Environments: []Environment{
			{ID: "prod"},
			{ID: "dev"},
		},
	}
	if err := b.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBlueprintValidateMissingServiceID(t *testing.T) {
	b := &Blueprint{Services: []Ref{{Ref: "ns/x"}}}
	err := b.Validate()
	if !errors.Is(err, ErrMissingServiceID) {
		t.Fatalf("err = %v, want ErrMissingServiceID", err)
	}
}

func TestBlueprintValidateDuplicateServiceID(t *testing.T) {
	b := &Blueprint{Services: []Ref{
		{ID: "x", Ref: "ns/a"},
		{ID: "x", Ref: "ns/b"},
	}}
	err := b.Validate()
	if !errors.Is(err, ErrDuplicateServiceID) {
		t.Fatalf("err = %v, want ErrDuplicateServiceID", err)
	}
}

func TestBlueprintValidateRouteUnknownService(t *testing.T) {
	b := &Blueprint{
		Services: []Ref{{ID: "api", Ref: "ns/api"}},
		Gateway:  Gateway{Routes: []Route{{Pattern: "/api", Service: "missing"}}},
	}
	err := b.Validate()
	if !errors.Is(err, ErrRouteServiceNotFound) {
		t.Fatalf("err = %v, want ErrRouteServiceNotFound", err)
	}
}

func TestBlueprintValidateDuplicateEnvironmentID(t *testing.T) {
	b := &Blueprint{
		Services:     []Ref{{ID: "api", Ref: "ns/api"}},
		Environments: []Environment{{ID: "prod"}, {ID: "prod"}},
	}
	err := b.Validate()
	if !errors.Is(err, ErrDuplicateEnvironmentID) {
		t.Fatalf("err = %v, want ErrDuplicateEnvironmentID", err)
	}
}

func TestBlueprintValidateBadServiceRef(t *testing.T) {
	b := &Blueprint{Services: []Ref{{ID: "x"}}}
	if err := b.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBlueprintValidateBadGateway(t *testing.T) {
	b := &Blueprint{
		Services: []Ref{{ID: "x", Ref: "ns/x"}},
		Gateway:  Gateway{Routes: []Route{{Service: "x"}}},
	}
	if err := b.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
