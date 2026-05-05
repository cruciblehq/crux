package manifest

import (
	"errors"
	"testing"
)

func TestServiceValidateOK(t *testing.T) {
	s := &Service{
		Recipe:     Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}},
		Entrypoint: []string{"/bin/run"},
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestServiceValidateMissingEntrypoint(t *testing.T) {
	s := &Service{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}}
	err := s.Validate()
	if !errors.Is(err, ErrMissingEntrypoint) {
		t.Fatalf("err = %v, want ErrMissingEntrypoint", err)
	}
}

func TestServiceValidatePropagatesRecipeError(t *testing.T) {
	s := &Service{Entrypoint: []string{"/bin/run"}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
