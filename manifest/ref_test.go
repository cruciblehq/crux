package manifest

import (
	"errors"
	"testing"
)

func TestRefValidateOK(t *testing.T) {
	r := &Ref{Ref: "cruciblehq/foo"}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRefValidateMissingTarget(t *testing.T) {
	r := &Ref{}
	err := r.Validate()
	if !errors.Is(err, ErrMissingRefTarget) {
		t.Fatalf("err = %v, want ErrMissingRefTarget", err)
	}
}

func TestRefValidateAcceptsArgsOnly(t *testing.T) {
	r := &Ref{Ref: "x", Args: map[string]string{"k": "v"}}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRefValidateInvalidID(t *testing.T) {
	r := &Ref{Ref: "x", ID: "My-Service"}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidRefID) {
		t.Fatalf("err = %v, want ErrInvalidRefID", err)
	}
}

func TestRefValidateValidID(t *testing.T) {
	r := &Ref{Ref: "x", ID: "api"}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRefValidateInvalidArgKey(t *testing.T) {
	r := &Ref{Ref: "x", Args: map[string]string{"Bad-Key": "v"}}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidArgKey) {
		t.Fatalf("err = %v, want ErrInvalidArgKey", err)
	}
}
