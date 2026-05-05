package manifest

import (
	"errors"
	"testing"
)

func TestComputeValidateOK(t *testing.T) {
	c := &Compute{ID: "c1", Provider: "local"}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestComputeValidateMissingID(t *testing.T) {
	err := (&Compute{Provider: "local"}).Validate()
	if !errors.Is(err, ErrMissingComputeID) {
		t.Fatalf("err = %v, want ErrMissingComputeID", err)
	}
}

func TestComputeValidateMissingProvider(t *testing.T) {
	err := (&Compute{ID: "c1"}).Validate()
	if !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("err = %v, want ErrMissingProvider", err)
	}
}
