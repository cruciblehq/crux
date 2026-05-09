package reference

import "testing"

func TestParse(t *testing.T) {
	ref, err := Parse("namespace/name 1.0.0", "template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Type() != "template" {
		t.Errorf("expected type %q, got %q", "template", ref.Type())
	}
	if ref.Namespace() != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", ref.Namespace())
	}
	if ref.Name() != "name" {
		t.Errorf("expected name %q, got %q", "name", ref.Name())
	}
}

func TestParse_BareName(t *testing.T) {
	ref, err := Parse("widget 1.0.0", "template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Namespace() != "" {
		t.Errorf("expected empty namespace, got %q", ref.Namespace())
	}
	if ref.Name() != "widget" {
		t.Errorf("expected name %q, got %q", "widget", ref.Name())
	}
}

func TestParse_RegistryNamespaceName(t *testing.T) {
	ref, err := Parse("hub.example.com/namespace/name 1.0.0", "template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Registry() != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", ref.Registry())
	}
	if ref.Namespace() != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", ref.Namespace())
	}
	if ref.Name() != "name" {
		t.Errorf("expected name %q, got %q", "name", ref.Name())
	}
}

func TestParse_Error(t *testing.T) {
	_, err := Parse("", "template")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMustParse(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0", "template")

	if ref.Name() != "name" {
		t.Errorf("expected name %q, got %q", "name", ref.Name())
	}
}

func TestMustParse_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	MustParse("", "template")
}

func TestReference_WithDefaults(t *testing.T) {
	ref, err := Parse("widget 1.0.0", "template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Registry() != "" {
		t.Errorf("expected empty registry, got %q", ref.Registry())
	}
	if ref.Namespace() != "" {
		t.Errorf("expected empty namespace, got %q", ref.Namespace())
	}

	resolved := ref.WithDefaults("hub.example.com", "official")

	if resolved.Registry() != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", resolved.Registry())
	}
	if resolved.Namespace() != "official" {
		t.Errorf("expected namespace %q, got %q", "official", resolved.Namespace())
	}

	// Original should be unchanged.
	if ref.Registry() != "" {
		t.Errorf("original registry should still be empty, got %q", ref.Registry())
	}
	if ref.Namespace() != "" {
		t.Errorf("original namespace should still be empty, got %q", ref.Namespace())
	}
}

func TestReference_WithDefaults_NoOverwrite(t *testing.T) {
	ref, err := Parse("hub.example.com/myteam/widget 1.0.0", "template")
	if err != nil {
		t.Fatal(err)
	}

	resolved := ref.WithDefaults("other.registry.com", "other-namespace")

	if resolved.Registry() != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", resolved.Registry())
	}
	if resolved.Namespace() != "myteam" {
		t.Errorf("expected namespace %q, got %q", "myteam", resolved.Namespace())
	}
}

func TestReference_Version(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0", "template")

	if ref.Version() == nil {
		t.Fatal("expected version, got nil")
	}
}

func TestReference_Channel(t *testing.T) {
	ref := MustParse("namespace/name :stable", "template")

	if ref.Channel() == nil {
		t.Fatal("expected channel, got nil")
	}
	if *ref.Channel() != "stable" {
		t.Errorf("expected channel %q, got %q", "stable", *ref.Channel())
	}
}

func TestReference_Digest(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0 sha256:abcd1234", "template")

	if ref.Digest() == nil {
		t.Fatal("expected digest, got nil")
	}
}

func TestReference_IsFrozen_True(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0 sha256:abcd1234", "template")

	if !ref.IsFrozen() {
		t.Error("expected IsFrozen to be true")
	}
}

func TestReference_IsFrozen_False(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0", "template")

	if ref.IsFrozen() {
		t.Error("expected IsFrozen to be false")
	}
}

func TestReference_IsChannelBased_True(t *testing.T) {
	ref := MustParse("namespace/name :stable", "template")

	if !ref.IsChannelBased() {
		t.Error("expected IsChannelBased to be true")
	}
}

func TestReference_IsChannelBased_False(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0", "template")

	if ref.IsChannelBased() {
		t.Error("expected IsChannelBased to be false")
	}
}

func TestReference_IsVersionBased_True(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0", "template")

	if !ref.IsVersionBased() {
		t.Error("expected IsVersionBased to be true")
	}
}

func TestReference_IsVersionBased_False(t *testing.T) {
	ref := MustParse("namespace/name :stable", "template")

	if ref.IsVersionBased() {
		t.Error("expected IsVersionBased to be false")
	}
}

func TestReference_String_WithVersion(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0", "template")

	s := ref.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestReference_String_WithChannel(t *testing.T) {
	ref := MustParse("namespace/name :stable", "template")

	s := ref.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestReference_String_WithDigest(t *testing.T) {
	ref := MustParse("namespace/name 1.0.0 sha256:abcd1234", "template")

	s := ref.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
