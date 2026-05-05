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
	if !errors.Is(err, ErrMissingRoutePattern) {
		t.Fatalf("err = %v, want ErrMissingRoutePattern", err)
	}
}

func TestRouteValidateMissingService(t *testing.T) {
	err := (&Route{Pattern: "/api"}).Validate()
	if !errors.Is(err, ErrMissingRouteService) {
		t.Fatalf("err = %v, want ErrMissingRouteService", err)
	}
}
