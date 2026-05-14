package manifest

import (
	"errors"
	"testing"
)

func TestRouteValidateOK(t *testing.T) {
	r := &Route{Pattern: "/api", Service: "svc"}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRouteValidateMissingPattern(t *testing.T) {
	err := (&Route{Service: "svc"}).Validate()
	if !errors.Is(err, ErrInvalidRoutePattern) {
		t.Fatalf("err = %v, want ErrInvalidRoutePattern", err)
	}
}

func TestRouteValidateMissingService(t *testing.T) {
	err := (&Route{Pattern: "/api"}).Validate()
	if !errors.Is(err, ErrInvalidRouteService) {
		t.Fatalf("err = %v, want ErrInvalidRouteService", err)
	}
}
