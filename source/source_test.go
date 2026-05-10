package source

import (
	"errors"
	"testing"
)

func TestNewSource_Valid(t *testing.T) {
	s, err := NewSource("https://hub.crucible.sh", "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Registry != "https://hub.crucible.sh" {
		t.Errorf("Registry = %q, want %q", s.Registry, "https://hub.crucible.sh")
	}
	if s.Namespace != "acme" {
		t.Errorf("Namespace = %q, want %q", s.Namespace, "acme")
	}
}

func TestNewSource_MissingRegistry(t *testing.T) {
	_, err := NewSource("", "acme")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingOption) {
		t.Errorf("expected ErrMissingOption in chain, got: %v", err)
	}
	if !errors.Is(err, ErrMissingRegistry) {
		t.Errorf("expected ErrMissingRegistry in chain, got: %v", err)
	}
}

func TestNewSource_MissingNamespace(t *testing.T) {
	_, err := NewSource("https://hub.crucible.sh", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingOption) {
		t.Errorf("expected ErrMissingOption in chain, got: %v", err)
	}
	if !errors.Is(err, ErrMissingNamespace) {
		t.Errorf("expected ErrMissingNamespace in chain, got: %v", err)
	}
}

func TestSource_Parse_AppliesDefaults(t *testing.T) {
	s, err := NewSource("https://hub.crucible.sh", "acme")
	if err != nil {
		t.Fatal(err)
	}

	// Reference with no registry or namespace — defaults should be filled in.
	ref, err := s.Parse("runtime", "myruntime 1.0.0")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if ref.Registry() != "https://hub.crucible.sh" {
		t.Errorf("Registry = %q, want %q", ref.Registry(), "https://hub.crucible.sh")
	}
	if ref.Namespace() != "acme" {
		t.Errorf("Namespace = %q, want %q", ref.Namespace(), "acme")
	}
	if ref.Name() != "myruntime" {
		t.Errorf("Name = %q, want %q", ref.Name(), "myruntime")
	}
}

func TestSource_Parse_ExplicitRegistryAndNamespace(t *testing.T) {
	s, err := NewSource("https://hub.crucible.sh", "acme")
	if err != nil {
		t.Fatal(err)
	}

	// Fully-qualified reference must not be overridden by defaults.
	ref, err := s.Parse("runtime", "other.registry.io/otherns/myruntime 1.2.3")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if ref.Registry() != "other.registry.io" {
		t.Errorf("Registry = %q, want %q", ref.Registry(), "other.registry.io")
	}
	if ref.Namespace() != "otherns" {
		t.Errorf("Namespace = %q, want %q", ref.Namespace(), "otherns")
	}
}

func TestSource_Parse_InvalidRef(t *testing.T) {
	s, err := NewSource("https://hub.crucible.sh", "acme")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Parse("runtime", ":::invalid:::")
	if err == nil {
		t.Fatal("expected error for invalid reference, got nil")
	}
}
