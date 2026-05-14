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

func TestServiceValidatePropagatesSchemaError(t *testing.T) {
	s := &Service{
		Schema:     &Schema{Default: "missing"},
		Entrypoint: []string{"/bin/run"},
		Recipe:     Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}},
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidService) {
		t.Fatalf("err = %v, want ErrInvalidService", err)
	}
}

func TestServiceEncodeWithSchema(t *testing.T) {
	s := &Service{
		Schema:     &Schema{Params: []Param{{Name: "port"}}},
		Entrypoint: []string{"/bin/run"},
		Recipe:     Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}},
	}
	raw, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("Encode() = %T, want map[string]any", raw)
	}
	if _, ok := m["schema"]; !ok {
		t.Fatal("expected schema key in encoded service")
	}
}

func TestServiceEncodeWithEntrypoint(t *testing.T) {
	s := &Service{
		Entrypoint: []string{"/bin/run", "--verbose"},
		Recipe:     Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}},
	}
	raw, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m := raw.(map[string]any)
	ep, ok := m["entrypoint"].([]string)
	if !ok {
		t.Fatalf("entrypoint = %T, want []string", m["entrypoint"])
	}
	if len(ep) != 2 || ep[0] != "/bin/run" {
		t.Fatalf("entrypoint = %v, want [\"/bin/run\", \"--verbose\"]", ep)
	}
}

func TestServiceEncodeWithoutSchema(t *testing.T) {
	s := &Service{
		Entrypoint: []string{"/bin/run"},
		Recipe:     Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}},
	}
	raw, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m := raw.(map[string]any)
	if _, ok := m["schema"]; ok {
		t.Fatal("expected no schema key for nil schema")
	}
}
