package net

import (
	"errors"
	"testing"
)

func TestSpecValidateAcceptsValidRules(t *testing.T) {
	s := &Spec{
		Ingress: []IngressRule{{Protocol: "tcp", Port: 80}},
		Egress:  []EgressRule{{Protocol: "tcp", Port: 443, Destination: "example.com"}},
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestSpecValidateAcceptsEmpty(t *testing.T) {
	if err := (&Spec{}).Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestSpecValidateRejectsBadIngress(t *testing.T) {
	s := &Spec{Ingress: []IngressRule{{Protocol: "bogus", Port: 80}}}
	if err := s.Validate(); !errors.Is(err, ErrInvalidSpec) {
		t.Fatalf("err = %v, want ErrInvalidSpec", err)
	}
}

func TestSpecValidateRejectsBadEgress(t *testing.T) {
	s := &Spec{Egress: []EgressRule{{Protocol: "tcp"}}}
	if err := s.Validate(); !errors.Is(err, ErrInvalidSpec) {
		t.Fatalf("err = %v, want ErrInvalidSpec", err)
	}
}
