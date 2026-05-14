package manifest

import (
	"errors"
	"testing"
)

func TestSchemaValidateEmpty(t *testing.T) {
	if err := (&Schema{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaValidateOK(t *testing.T) {
	s := &Schema{
		Default: "name",
		Params:  []Param{{Name: "name"}, {Name: "value", Default: "x"}},
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaValidateDuplicateParam(t *testing.T) {
	s := &Schema{Params: []Param{{Name: "x"}, {Name: "x"}}}
	err := s.Validate()
	if !errors.Is(err, ErrDuplicateParamName) {
		t.Fatalf("err = %v, want ErrDuplicateParamName", err)
	}
}

func TestSchemaValidateDefaultNotInSchema(t *testing.T) {
	s := &Schema{Default: "missing", Params: []Param{{Name: "other"}}}
	err := s.Validate()
	if !errors.Is(err, ErrDefaultNotInSchema) {
		t.Fatalf("err = %v, want ErrDefaultNotInSchema", err)
	}
}

func TestSchemaValidateDefaultInvalidFormat(t *testing.T) {
	s := &Schema{Default: "INVALID"}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidParamName) {
		t.Fatalf("err = %v, want ErrInvalidParamName", err)
	}
}

func TestSchemaValidatePropagatesParamError(t *testing.T) {
	s := &Schema{Params: []Param{{Name: ""}}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
