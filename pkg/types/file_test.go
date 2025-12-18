package types

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	v := testStruct{Name: "test", Version: 1, Enabled: true}

	if err := EncodeFile(path, "key", v); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestEncodeFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	v := testStruct{Name: "test", Version: 1, Enabled: true}

	if err := EncodeFile(path, "key", v); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestEncodeFile_YML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	v := testStruct{Name: "test", Version: 1, Enabled: true}

	if err := EncodeFile(path, "key", v); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestEncodeFile_TOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")

	v := testStruct{Name: "test", Version: 1, Enabled: true}

	if err := EncodeFile(path, "key", v); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestEncodeFile_UnknownExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.xml")

	v := testStruct{Name: "test", Version: 1, Enabled: true}

	err := EncodeFile(path, "key", v)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrUnsupportedContentType) {
		t.Errorf("expected ErrUnsupportedContentType, got %v", err)
	}
}

func TestEncodeFile_InvalidPath(t *testing.T) {
	path := "/nonexistent/directory/test.json"

	v := testStruct{Name: "test", Version: 1, Enabled: true}

	err := EncodeFile(path, "key", v)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrEncodingFailed) {
		t.Errorf("expected ErrEncodingFailed, got %v", err)
	}
}

func TestDecodeFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	if err := os.WriteFile(path, []byte(`{"name":"test","version":1,"enabled":true}`), 0644); err != nil {
		t.Fatal(err)
	}

	var target testStruct
	ct, err := DecodeFile(path, "key", &target)
	if err != nil {
		t.Fatal(err)
	}

	if ct != ContentTypeJSON {
		t.Errorf("expected ContentTypeJSON, got %v", ct)
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

func TestDecodeFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	if err := os.WriteFile(path, []byte("name: test\nversion: 1\nenabled: true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var target testStruct
	ct, err := DecodeFile(path, "key", &target)
	if err != nil {
		t.Fatal(err)
	}

	if ct != ContentTypeYAML {
		t.Errorf("expected ContentTypeYAML, got %v", ct)
	}
	if target.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", target.Name)
	}
}

func TestDecodeFile_YML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	if err := os.WriteFile(path, []byte("name: test\nversion: 1\nenabled: true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var target testStruct
	ct, err := DecodeFile(path, "key", &target)
	if err != nil {
		t.Fatal(err)
	}

	if ct != ContentTypeYAML {
		t.Errorf("expected ContentTypeYAML, got %v", ct)
	}
}

func TestDecodeFile_TOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")

	if err := os.WriteFile(path, []byte("name = \"test\"\nversion = 1\nenabled = true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var target testStruct
	ct, err := DecodeFile(path, "key", &target)
	if err != nil {
		t.Fatal(err)
	}

	if ct != ContentTypeTOML {
		t.Errorf("expected ContentTypeTOML, got %v", ct)
	}
	if target.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", target.Name)
	}
}

func TestDecodeFile_UnknownExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.xml")

	if err := os.WriteFile(path, []byte("<test/>"), 0644); err != nil {
		t.Fatal(err)
	}

	var target testStruct
	_, err := DecodeFile(path, "key", &target)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrUnsupportedContentType) {
		t.Errorf("expected ErrUnsupportedContentType, got %v", err)
	}
}

func TestDecodeFile_FileNotFound(t *testing.T) {
	path := "/nonexistent/file.json"

	var target testStruct
	_, err := DecodeFile(path, "key", &target)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrDecodingFailed) {
		t.Errorf("expected ErrDecodingFailed, got %v", err)
	}
}

func TestDecodeFile_InvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	if err := os.WriteFile(path, []byte(`{invalid}`), 0644); err != nil {
		t.Fatal(err)
	}

	var target testStruct
	_, err := DecodeFile(path, "key", &target)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrDecodingFailed) {
		t.Errorf("expected ErrDecodingFailed, got %v", err)
	}
}

func TestRoundtripFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	original := testStruct{Name: "test", Version: 42, Enabled: true}

	if err := EncodeFile(path, "key", original); err != nil {
		t.Fatal(err)
	}

	var decoded testStruct
	_, err := DecodeFile(path, "key", &decoded)
	if err != nil {
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

func TestRoundtripFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	original := testStruct{Name: "test", Version: 42, Enabled: true}

	if err := EncodeFile(path, "key", original); err != nil {
		t.Fatal(err)
	}

	var decoded testStruct
	_, err := DecodeFile(path, "key", &decoded)
	if err != nil {
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

func TestRoundtripFile_TOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")

	original := testStruct{Name: "test", Version: 42, Enabled: true}

	if err := EncodeFile(path, "key", original); err != nil {
		t.Fatal(err)
	}

	var decoded testStruct
	_, err := DecodeFile(path, "key", &decoded)
	if err != nil {
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

func TestContentTypeFromExtension_CaseInsensitive(t *testing.T) {
	ct, err := contentTypeFromExtension("test.JSON")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeJSON {
		t.Errorf("expected ContentTypeJSON, got %v", ct)
	}
}

func TestContentTypeFromExtension_NoExtension(t *testing.T) {
	_, err := contentTypeFromExtension("test")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrUnsupportedContentType) {
		t.Errorf("expected ErrUnsupportedContentType, got %v", err)
	}
}
