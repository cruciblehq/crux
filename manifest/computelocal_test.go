package manifest

import (
	"errors"
	"testing"
)

func TestComputeValidateOK(t *testing.T) {
	c := &Compute{Provider: "local"}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestComputeValidateMissingProvider(t *testing.T) {
	err := (&Compute{}).Validate()
	if !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("err = %v, want ErrMissingProvider", err)
	}
}

func TestComputeValidateUnknownProvider(t *testing.T) {
	err := (&Compute{Provider: "gcp"}).Validate()
	if !errors.Is(err, ErrInvalidProviderType) {
		t.Fatalf("err = %v, want ErrInvalidProviderType", err)
	}
}
