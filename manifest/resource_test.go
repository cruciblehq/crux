package manifest

import (
	"errors"
	"testing"
)

func TestResourceValidateOK(t *testing.T) {
	r := &Resource{Type: TypeService, Name: "ns/x", Version: "1.0.0"}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestResourceValidateMissingName(t *testing.T) {
	r := &Resource{Type: TypeService, Version: "1.0.0"}
	err := r.Validate()
	if !errors.Is(err, ErrMissingName) {
		t.Fatalf("err = %v, want ErrMissingName", err)
	}
}

func TestResourceValidateMissingVersion(t *testing.T) {
	r := &Resource{Type: TypeService, Name: "ns/x"}
	err := r.Validate()
	if !errors.Is(err, ErrMissingVersion) {
		t.Fatalf("err = %v, want ErrMissingVersion", err)
	}
}

func TestResourceValidateInvalidType(t *testing.T) {
	r := &Resource{Type: "unknown", Name: "ns/x", Version: "1.0.0"}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResourceType) {
		t.Fatalf("err = %v, want ErrInvalidResourceType", err)
	}
}

func TestResourceValidateEmptyType(t *testing.T) {
	r := &Resource{Name: "ns/x", Version: "1.0.0"}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResourceType) {
		t.Fatalf("err = %v, want ErrInvalidResourceType", err)
	}
}
