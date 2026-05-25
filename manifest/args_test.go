package manifest

import (
	"errors"
	"testing"
)

func TestArgsValidateValid(t *testing.T) {
	a := Args{"foo": "bar", "baz": "qux"}
	if err := a.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestArgsValidateEmpty(t *testing.T) {
	var a Args
	if err := a.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestArgsValidateInvalidKey(t *testing.T) {
	a := Args{"bad key!": "value"}
	err := a.Validate()
	if !errors.Is(err, ErrInvalidArgKey) {
		t.Fatalf("Validate() = %v, want ErrInvalidArgKey", err)
	}
}
