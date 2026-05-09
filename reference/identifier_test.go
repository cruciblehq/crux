package reference

import "testing"

func TestParseIdentifier(t *testing.T) {
	id, err := ParseIdentifier("namespace/name", "template")
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

func TestParseIdentifier_BareName(t *testing.T) {
	id, err := ParseIdentifier("widget", "template")
	if err != nil {
		t.Fatal(err)
	}

	if id.Registry() != "" {
		t.Errorf("expected empty registry, got %q", id.Registry())
	}
	if id.Namespace() != "" {
		t.Errorf("expected empty namespace, got %q", id.Namespace())
	}
	if id.Name() != "widget" {
		t.Errorf("expected name %q, got %q", "widget", id.Name())
	}
}

func TestParseIdentifier_RegistryNamespaceName(t *testing.T) {
	id, err := ParseIdentifier("hub.example.com/namespace/name", "template")
	if err != nil {
		t.Fatal(err)
	}

	if id.Registry() != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", id.Registry())
	}
	if id.Namespace() != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", id.Namespace())
	}
	if id.Name() != "name" {
		t.Errorf("expected name %q, got %q", "name", id.Name())
	}
}

func TestParseIdentifier_Error(t *testing.T) {
	_, err := ParseIdentifier("", "template")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMustParseIdentifier(t *testing.T) {
	id := MustParseIdentifier("namespace/name", "template")

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

	MustParseIdentifier("", "template")
}

func TestIdentifier_WithDefaults(t *testing.T) {
	id, err := ParseIdentifier("widget", "template")
	if err != nil {
		t.Fatal(err)
	}

	if id.Registry() != "" {
		t.Errorf("expected empty registry, got %q", id.Registry())
	}
	if id.Namespace() != "" {
		t.Errorf("expected empty namespace, got %q", id.Namespace())
	}

	resolved := id.WithDefaults("hub.example.com", "official")

	if resolved.Registry() != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", resolved.Registry())
	}
	if resolved.Namespace() != "official" {
		t.Errorf("expected namespace %q, got %q", "official", resolved.Namespace())
	}

	// Original should be unchanged.
	if id.Registry() != "" {
		t.Errorf("original registry should still be empty, got %q", id.Registry())
	}
	if id.Namespace() != "" {
		t.Errorf("original namespace should still be empty, got %q", id.Namespace())
	}
}

func TestIdentifier_WithDefaults_NoOverwrite(t *testing.T) {
	id, err := ParseIdentifier("hub.example.com/myteam/widget", "template")
	if err != nil {
		t.Fatal(err)
	}

	resolved := id.WithDefaults("other.registry.com", "other-namespace")

	if resolved.Registry() != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", resolved.Registry())
	}
	if resolved.Namespace() != "myteam" {
		t.Errorf("expected namespace %q, got %q", "myteam", resolved.Namespace())
	}
}

func TestIdentifier_Path_NamespaceAndName(t *testing.T) {
	id := MustParseIdentifier("namespace/name", "template")

	if id.Path() != "namespace/name" {
		t.Errorf("expected path %q, got %q", "namespace/name", id.Path())
	}
}

func TestIdentifier_Path_NameOnly(t *testing.T) {
	id := MustParseIdentifier("widget", "template")

	if id.Path() != "widget" {
		t.Errorf("expected path %q, got %q", "widget", id.Path())
	}
}

func TestIdentifier_String_WithRegistry(t *testing.T) {
	id, err := ParseIdentifier("hub.example.com/namespace/name", "template")
	if err != nil {
		t.Fatal(err)
	}

	expected := "template hub.example.com/namespace/name"
	if id.String() != expected {
		t.Errorf("expected string %q, got %q", expected, id.String())
	}
}

func TestIdentifier_String_WithoutRegistry(t *testing.T) {
	id := MustParseIdentifier("namespace/name", "template")

	expected := "template namespace/name"
	if id.String() != expected {
		t.Errorf("expected string %q, got %q", expected, id.String())
	}
}
