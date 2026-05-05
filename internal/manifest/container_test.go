package manifest

import (
	"errors"
	"testing"
)

func TestContainerValidateOK(t *testing.T) {
	c := &Container{Service: "s", Compute: "c"}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestContainerValidateMissingService(t *testing.T) {
	err := (&Container{Compute: "c"}).Validate()
	if !errors.Is(err, ErrMissingContainerService) {
		t.Fatalf("err = %v, want ErrMissingContainerService", err)
	}
}

func TestContainerValidateMissingCompute(t *testing.T) {
	err := (&Container{Service: "s"}).Validate()
	if !errors.Is(err, ErrMissingContainerCompute) {
		t.Fatalf("err = %v, want ErrMissingContainerCompute", err)
	}
}
