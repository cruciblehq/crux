package reference

import (
	"testing"
)

func TestParseIdentifier(t *testing.T) {
	id, err := ParseIdentifier("namespace/name", "template", nil)
	if err != nil {
		t.Fatal(err)
	}

	if id.Type() != "template" {
		t.Errorf("expected type %q, got %q", "template", id.Type())
	}
	if id.Namespace() != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", id.Namespace())
	}
	if id.Name() != "name" {
		t.Errorf("expected name %q, got %q", "name", id.Name())
	}
}

func TestParseIdentifier_WithOptions(t *testing.T) {
	opts := &IdentifierOptions{
		DefaultScheme:    "oci",
		DefaultRegistry:  "custom.registry.io",
		DefaultNamespace: "crucible",
	}

	id, err := ParseIdentifier("widget", "template", opts)
	if err != nil {
		t.Fatal(err)
	}

	if id.Scheme() != "oci" {
		t.Errorf("expected scheme %q, got %q", "oci", id.Scheme())
	}
	if id.Registry() != "custom.registry.io" {
		t.Errorf("expected registry %q, got %q", "custom.registry.io", id.Registry())
	}
	if id.Namespace() != "crucible" {
		t.Errorf("expected namespace %q, got %q", "crucible", id.Namespace())
	}
}

func TestParseIdentifier_Error(t *testing.T) {
	_, err := ParseIdentifier("", "template", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMustParseIdentifier(t *testing.T) {
	id := MustParseIdentifier("namespace/name", "template", nil)

	if id.Name() != "name" {
		t.Errorf("expected name %q, got %q", "name", id.Name())
	}
}

func TestMustParseIdentifier_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	MustParseIdentifier("", "template", nil)
}

func TestIdentifier_Path_DefaultRegistry(t *testing.T) {
	id := MustParseIdentifier("namespace/name", "template", nil)

	if id.Path() != "namespace/name" {
		t.Errorf("expected path %q, got %q", "namespace/name", id.Path())
	}
}

func TestIdentifier_Path_CustomRegistry(t *testing.T) {
	id := MustParseIdentifier("myregistry.com/path/to/resource", "template", nil)

	if id.Path() != "path/to/resource" {
		t.Errorf("expected path %q, got %q", "path/to/resource", id.Path())
	}
}

func TestIdentifier_URI(t *testing.T) {
	id := MustParseIdentifier("namespace/name", "template", nil)

	expected := "https://registry.crucible.net/namespace/name"
	if id.URI() != expected {
		t.Errorf("expected URI %q, got %q", expected, id.URI())
	}
}

func TestIdentifier_String(t *testing.T) {
	id := MustParseIdentifier("namespace/name", "template", nil)

	expected := "template https://registry.crucible.net/namespace/name"
	if id.String() != expected {
		t.Errorf("expected string %q, got %q", expected, id.String())
	}
}

func TestIdentifierOptions_NilReceiver(t *testing.T) {
	var opts *IdentifierOptions

	if opts.scheme() != DefaultScheme {
		t.Errorf("expected scheme %q, got %q", DefaultScheme, opts.scheme())
	}
	if opts.registry() != DefaultRegistry {
		t.Errorf("expected registry %q, got %q", DefaultRegistry, opts.registry())
	}
	if opts.namespace() != DefaultNamespace {
		t.Errorf("expected namespace %q, got %q", DefaultNamespace, opts.namespace())
	}
}

func TestIdentifierOptions_EmptyFields(t *testing.T) {
	opts := &IdentifierOptions{}

	if opts.scheme() != DefaultScheme {
		t.Errorf("expected scheme %q, got %q", DefaultScheme, opts.scheme())
	}
	if opts.registry() != DefaultRegistry {
		t.Errorf("expected registry %q, got %q", DefaultRegistry, opts.registry())
	}
	if opts.namespace() != DefaultNamespace {
		t.Errorf("expected namespace %q, got %q", DefaultNamespace, opts.namespace())
	}
}

func TestIdentifierOptions_PartialFields(t *testing.T) {
	opts := &IdentifierOptions{
		DefaultNamespace: "custom",
	}

	if opts.scheme() != DefaultScheme {
		t.Errorf("expected scheme %q, got %q", DefaultScheme, opts.scheme())
	}
	if opts.registry() != DefaultRegistry {
		t.Errorf("expected registry %q, got %q", DefaultRegistry, opts.registry())
	}
	if opts.namespace() != "custom" {
		t.Errorf("expected namespace %q, got %q", "custom", opts.namespace())
	}
}
