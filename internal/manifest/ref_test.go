package manifest

import (
	"errors"
	"testing"
)

func TestRefValidateOK(t *testing.T) {
	r := &Ref{Target: "cruciblehq/foo"}
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

func TestRefValidateRejectsValueAndArgs(t *testing.T) {
	r := &Ref{Target: "x", Value: "v", Args: map[string]string{"k": "v"}}
	err := r.Validate()
	if !errors.Is(err, ErrRefMixed) {
		t.Fatalf("err = %v, want ErrRefMixed", err)
	}
}

func TestRefValidateAcceptsValueOnly(t *testing.T) {
	r := &Ref{Target: "x", Value: "v"}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRefValidateAcceptsArgsOnly(t *testing.T) {
	r := &Ref{Target: "x", Args: map[string]string{"k": "v"}}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}
