package codec

import (
	"errors"
	"testing"
)

func TestEncode_JSON(t *testing.T) {
	s := sample{Name: "foo", Version: 1}
	data, err := Encode(s, JSON)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if got != `{"name":"foo","version":1}` {
		t.Errorf("Encode(JSON) = %s", got)
	}
}

func TestEncode_YAML(t *testing.T) {
	s := sample{Name: "foo", Version: 1}
	data, err := Encode(s, YAML)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	want := "name: foo\nversion: 1\n"
	if got != want {
		t.Errorf("Encode(YAML) = %q, want %q", got, want)
	}
}

func TestEncode_UnsupportedFormat(t *testing.T) {
	_, err := Encode(sample{}, Format(99))
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Encode(Format(99)) error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestEncode_CustomEncodable(t *testing.T) {
	c := &custom{Value: "hello"}
	data, err := Encode(c, JSON)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"custom":"hello"}` {
		t.Errorf("Encode(custom) = %q", data)
	}
}

func TestToMap(t *testing.T) {
	s := sample{Name: "x", Version: 5}
	m, err := ToMap(s)
	if err != nil {
		t.Fatal(err)
	}
	if m["name"] != "x" {
		t.Errorf("ToMap name = %v, want %q", m["name"], "x")
	}
	if m["version"] != 5 {
		t.Errorf("ToMap version = %v, want 5", m["version"])
	}
}

func TestToMap_Nested(t *testing.T) {
	n := nested{Inner: sample{Name: "n", Version: 1}}
	m, err := ToMap(n)
	if err != nil {
		t.Fatal(err)
	}
	inner, ok := m["inner"].(map[string]any)
	if !ok {
		t.Fatalf("ToMap inner = %T, want map[string]any", m["inner"])
	}
	if inner["name"] != "n" {
		t.Errorf("ToMap inner.name = %v, want %q", inner["name"], "n")
	}
}

func TestToMap_Squash(t *testing.T) {
	s := squashed{sample: sample{Name: "sq", Version: 7}, Extra: "e"}
	m, err := ToMap(s)
	if err != nil {
		t.Fatal(err)
	}
	if m["extra"] != "e" {
		t.Errorf("ToMap squash extra = %v, want %q", m["extra"], "e")
	}
}

func TestRoundtrip_JSON(t *testing.T) {
	orig := sample{Name: "rt", Version: 42}
	data, err := Encode(orig, JSON)
	if err != nil {
		t.Fatal(err)
	}
	var got sample
	if err := Unmarshal(data, &got, JSON); err != nil {
		t.Fatal(err)
	}
	if got != orig {
		t.Errorf("roundtrip JSON: got %+v, want %+v", got, orig)
	}
}

func TestRoundtrip_YAML(t *testing.T) {
	orig := sample{Name: "rt", Version: 42}
	data, err := Encode(orig, YAML)
	if err != nil {
		t.Fatal(err)
	}
	var got sample
	if err := Unmarshal(data, &got, YAML); err != nil {
		t.Fatal(err)
	}
	if got != orig {
		t.Errorf("roundtrip YAML: got %+v, want %+v", got, orig)
	}
}
