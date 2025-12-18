package types

import (
	"errors"
	"strings"
	"testing"
)

type testStruct struct {
	Name    string `key:"name"`
	Version int    `key:"version"`
	Enabled bool   `key:"enabled"`
}

func TestEncode_JSON(t *testing.T) {
	v := testStruct{Name: "test", Version: 1, Enabled: true}

	data, err := Encode(ContentTypeJSON, "key", v)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `"name"`) {
		t.Error("expected JSON to contain 'name' key")
	}
	if !strings.Contains(string(data), `"test"`) {
		t.Error("expected JSON to contain 'test' value")
	}
}

func TestEncode_YAML(t *testing.T) {
	v := testStruct{Name: "test", Version: 1, Enabled: true}

	data, err := Encode(ContentTypeYAML, "key", v)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "name:") {
		t.Error("expected YAML to contain 'name' key")
	}
}

func TestEncode_TOML(t *testing.T) {
	v := testStruct{Name: "test", Version: 1, Enabled: true}

	data, err := Encode(ContentTypeTOML, "key", v)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "name =") {
		t.Error("expected TOML to contain 'name' key")
	}
}

func TestEncode_UnsupportedContentType(t *testing.T) {
	v := testStruct{Name: "test", Version: 1, Enabled: true}

	_, err := Encode(ContentTypeUnknown, "key", v)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrUnsupportedContentType) {
		t.Errorf("expected ErrUnsupportedContentType, got %v", err)
	}
}

func TestEncode_CustomTag(t *testing.T) {
	type customStruct struct {
		Name string `custom:"custom_name"`
	}

	v := customStruct{Name: "test"}

	data, err := Encode(ContentTypeJSON, "custom", v)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `"custom_name"`) {
		t.Error("expected JSON to contain 'custom_name' key")
	}
}

func TestDecode_JSON(t *testing.T) {
	data := []byte(`{"name":"test","version":1,"enabled":true}`)

	var target testStruct
	if err := Decode(ContentTypeJSON, "key", &target, data); err != nil {
		t.Fatal(err)
	}

	if target.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", target.Name)
	}
	if target.Version != 1 {
		t.Errorf("expected version %d, got %d", 1, target.Version)
	}
	if !target.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestDecode_YAML(t *testing.T) {
	data := []byte("name: test\nversion: 1\nenabled: true\n")

	var target testStruct
	if err := Decode(ContentTypeYAML, "key", &target, data); err != nil {
		t.Fatal(err)
	}

	if target.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", target.Name)
	}
	if target.Version != 1 {
		t.Errorf("expected version %d, got %d", 1, target.Version)
	}
	if !target.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestDecode_TOML(t *testing.T) {
	data := []byte("name = \"test\"\nversion = 1\nenabled = true\n")

	var target testStruct
	if err := Decode(ContentTypeTOML, "key", &target, data); err != nil {
		t.Fatal(err)
	}

	if target.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", target.Name)
	}
	if target.Version != 1 {
		t.Errorf("expected version %d, got %d", 1, target.Version)
	}
	if !target.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestDecode_UnsupportedContentType(t *testing.T) {
	data := []byte(`{"name":"test"}`)

	var target testStruct
	err := Decode(ContentTypeUnknown, "key", &target, data)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrUnsupportedContentType) {
		t.Errorf("expected ErrUnsupportedContentType, got %v", err)
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid}`)

	var target testStruct
	err := Decode(ContentTypeJSON, "key", &target, data)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrDecodingFailed) {
		t.Errorf("expected ErrDecodingFailed, got %v", err)
	}
}

func TestDecode_InvalidYAML(t *testing.T) {
	data := []byte(":\ninvalid")

	var target testStruct
	err := Decode(ContentTypeYAML, "key", &target, data)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrDecodingFailed) {
		t.Errorf("expected ErrDecodingFailed, got %v", err)
	}
}

func TestDecode_InvalidTOML(t *testing.T) {
	data := []byte("= invalid")

	var target testStruct
	err := Decode(ContentTypeTOML, "key", &target, data)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrDecodingFailed) {
		t.Errorf("expected ErrDecodingFailed, got %v", err)
	}
}

func TestDecode_CustomTag(t *testing.T) {
	type customStruct struct {
		Name string `custom:"custom_name"`
	}

	data := []byte(`{"custom_name":"test"}`)

	var target customStruct
	if err := Decode(ContentTypeJSON, "custom", &target, data); err != nil {
		t.Fatal(err)
	}

	if target.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", target.Name)
	}
}

func TestRoundtrip_JSON(t *testing.T) {
	original := testStruct{Name: "test", Version: 42, Enabled: true}

	data, err := Encode(ContentTypeJSON, "key", original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded testStruct
	if err := Decode(ContentTypeJSON, "key", &decoded, data); err != nil {
		t.Fatal(err)
	}

	if decoded.Name != original.Name {
		t.Errorf("expected name %q, got %q", original.Name, decoded.Name)
	}
	if decoded.Version != original.Version {
		t.Errorf("expected version %d, got %d", original.Version, decoded.Version)
	}
	if decoded.Enabled != original.Enabled {
		t.Errorf("expected enabled %v, got %v", original.Enabled, decoded.Enabled)
	}
}

func TestRoundtrip_YAML(t *testing.T) {
	original := testStruct{Name: "test", Version: 42, Enabled: true}

	data, err := Encode(ContentTypeYAML, "key", original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded testStruct
	if err := Decode(ContentTypeYAML, "key", &decoded, data); err != nil {
		t.Fatal(err)
	}

	if decoded.Name != original.Name {
		t.Errorf("expected name %q, got %q", original.Name, decoded.Name)
	}
	if decoded.Version != original.Version {
		t.Errorf("expected version %d, got %d", original.Version, decoded.Version)
	}
	if decoded.Enabled != original.Enabled {
		t.Errorf("expected enabled %v, got %v", original.Enabled, decoded.Enabled)
	}
}

func TestRoundtrip_TOML(t *testing.T) {
	original := testStruct{Name: "test", Version: 42, Enabled: true}

	data, err := Encode(ContentTypeTOML, "key", original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded testStruct
	if err := Decode(ContentTypeTOML, "key", &decoded, data); err != nil {
		t.Fatal(err)
	}

	if decoded.Name != original.Name {
		t.Errorf("expected name %q, got %q", original.Name, decoded.Name)
	}
	if decoded.Version != original.Version {
		t.Errorf("expected version %d, got %d", original.Version, decoded.Version)
	}
	if decoded.Enabled != original.Enabled {
		t.Errorf("expected enabled %v, got %v", original.Enabled, decoded.Enabled)
	}
}
