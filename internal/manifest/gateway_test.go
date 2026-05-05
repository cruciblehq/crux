package manifest

import (
	"errors"
	"testing"
)

func TestGatewayValidateEmpty(t *testing.T) {
	if err := (&Gateway{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestGatewayValidateOK(t *testing.T) {
	g := &Gateway{Routes: []Route{
		{Pattern: "/api", Service: "a"},
		{Pattern: "/auth", Service: "b"},
	}}
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestGatewayValidateDuplicatePattern(t *testing.T) {
	g := &Gateway{Routes: []Route{
		{Pattern: "/x", Service: "a"},
		{Pattern: "/x", Service: "b"},
	}}
	err := g.Validate()
	if !errors.Is(err, ErrDuplicateRoutePattern) {
		t.Fatalf("err = %v, want ErrDuplicateRoutePattern", err)
	}
}

func TestGatewayValidatePropagatesRouteError(t *testing.T) {
	g := &Gateway{Routes: []Route{{Pattern: ""}}}
	if err := g.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
