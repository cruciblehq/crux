package types

import (
	"errors"
	"testing"
)

func TestParseContentType_ApplicationJSON(t *testing.T) {
	ct, err := ParseContentType("application/json")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeJSON {
		t.Errorf("expected ContentTypeJSON, got %v", ct)
	}
}

func TestParseContentType_ApplicationJSONWithCharset(t *testing.T) {
	ct, err := ParseContentType("application/json; charset=utf-8")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeJSON {
		t.Errorf("expected ContentTypeJSON, got %v", ct)
	}
}

func TestParseContentType_TextJSON(t *testing.T) {
	ct, err := ParseContentType("text/json")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeJSON {
		t.Errorf("expected ContentTypeJSON, got %v", ct)
	}
}

func TestParseContentType_ApplicationYAML(t *testing.T) {
	ct, err := ParseContentType("application/yaml")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeYAML {
		t.Errorf("expected ContentTypeYAML, got %v", ct)
	}
}

func TestParseContentType_ApplicationXYAML(t *testing.T) {
	ct, err := ParseContentType("application/x-yaml")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeYAML {
		t.Errorf("expected ContentTypeYAML, got %v", ct)
	}
}

func TestParseContentType_TextYAML(t *testing.T) {
	ct, err := ParseContentType("text/yaml")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeYAML {
		t.Errorf("expected ContentTypeYAML, got %v", ct)
	}
}

func TestParseContentType_ApplicationTOML(t *testing.T) {
	ct, err := ParseContentType("application/toml")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeTOML {
		t.Errorf("expected ContentTypeTOML, got %v", ct)
	}
}

func TestParseContentType_ApplicationXTOML(t *testing.T) {
	ct, err := ParseContentType("application/x-toml")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeTOML {
		t.Errorf("expected ContentTypeTOML, got %v", ct)
	}
}

func TestParseContentType_TextTOML(t *testing.T) {
	ct, err := ParseContentType("text/toml")
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentTypeTOML {
		t.Errorf("expected ContentTypeTOML, got %v", ct)
	}
}

func TestParseContentType_Unsupported(t *testing.T) {
	_, err := ParseContentType("application/xml")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUnsupportedContentType) {
		t.Errorf("expected ErrUnsupportedContentType, got %v", err)
	}
}

func TestParseContentType_Invalid(t *testing.T) {
	_, err := ParseContentType("not a media type")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidContentType) {
		t.Errorf("expected ErrInvalidContentType, got %v", err)
	}
}

func TestParseContentType_Empty(t *testing.T) {
	_, err := ParseContentType("")
	if err == nil {
		t.Fatal("expected error")
	}
}
