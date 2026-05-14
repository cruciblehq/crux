package manifest

import (
	"errors"
	"testing"
)

func TestEnvironmentValidateOK(t *testing.T) {
	e := &Environment{ID: "prod"}
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentValidateMissingID(t *testing.T) {
	err := (&Environment{}).Validate()
	if !errors.Is(err, ErrMissingEnvironmentID) {
		t.Fatalf("err = %v, want ErrMissingEnvironmentID", err)
	}
}

func TestEnvironmentValidateInvalidID(t *testing.T) {
	err := (&Environment{ID: "PROD"}).Validate()
	if !errors.Is(err, ErrInvalidEnvironmentID) {
		t.Fatalf("err = %v, want ErrInvalidEnvironmentID", err)
	}
}
