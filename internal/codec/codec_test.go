package codec

import (
	"errors"
	"testing"
)

type sample struct {
	Name    string `codec:"name"`
	Version int    `codec:"version"`
}

type nested struct {
	Inner sample `codec:"inner"`
}

type squashed struct {
	sample `codec:",squash"`
	Extra  string `codec:"extra"`
}

type custom struct {
	Value string
}

func (c *custom) Encode(f Format) ([]byte, error) {
	return []byte("custom:" + c.Value), nil
}

func (c *custom) Decode(data []byte, f Format) error {
	c.Value = string(data)
	return nil
}

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

func TestEncode_CustomEncoder(t *testing.T) {
	c := &custom{Value: "hello"}
	data, err := Encode(c, JSON)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "custom:hello" {
		t.Errorf("Encode(custom) = %q, want %q", data, "custom:hello")
	}
}

func TestDecode_JSON(t *testing.T) {
	var s sample
	err := Decode([]byte(`{"name":"bar","version":2}`), &s, JSON)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "bar" || s.Version != 2 {
		t.Errorf("Decode(JSON) = %+v", s)
	}
}

func TestDecode_YAML(t *testing.T) {
	var s sample
	err := Decode([]byte("name: baz\nversion: 3\n"), &s, YAML)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "baz" || s.Version != 3 {
		t.Errorf("Decode(YAML) = %+v", s)
	}
}

func TestDecode_UnsupportedFormat(t *testing.T) {
	var s sample
	err := Decode([]byte("{}"), &s, Format(99))
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Decode(Format(99)) error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestDecode_InvalidData(t *testing.T) {
	var s sample
	err := Decode([]byte("not json"), &s, JSON)
	if err == nil {
		t.Error("Decode(invalid JSON) should fail")
	}
}

func TestDecode_CustomDecoder(t *testing.T) {
	c := &custom{}
	err := Decode([]byte("raw-input"), c, JSON)
	if err != nil {
		t.Fatal(err)
	}
	if c.Value != "raw-input" {
		t.Errorf("Decode(custom) = %q, want %q", c.Value, "raw-input")
	}
}

func TestRoundtrip_JSON(t *testing.T) {
	orig := sample{Name: "rt", Version: 42}
	data, err := Encode(orig, JSON)
	if err != nil {
		t.Fatal(err)
	}
	var got sample
	if err := Decode(data, &got, JSON); err != nil {
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
	if err := Decode(data, &got, YAML); err != nil {
		t.Fatal(err)
	}
	if got != orig {
		t.Errorf("roundtrip YAML: got %+v, want %+v", got, orig)
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

func TestFromMap_Squash(t *testing.T) {
	m := map[string]any{"name": "sq", "version": 7, "extra": "e"}
	var s squashed
	if err := FromMap(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "sq" {
		t.Errorf("FromMap squash name = %q, want %q", s.Name, "sq")
	}
	if s.Version != 7 {
		t.Errorf("FromMap squash version = %d, want 7", s.Version)
	}
	if s.Extra != "e" {
		t.Errorf("FromMap squash extra = %q, want %q", s.Extra, "e")
	}
}

func TestFromMap(t *testing.T) {
	m := map[string]any{"name": "y", "version": 10}
	var s sample
	if err := FromMap(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "y" || s.Version != 10 {
		t.Errorf("FromMap = %+v", s)
	}
}

func TestFromMap_WeaklyTyped(t *testing.T) {
	m := map[string]any{"name": "w", "version": "8"}
	var s sample
	if err := FromMap(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Version != 8 {
		t.Errorf("FromMap weakly typed version = %d, want 8", s.Version)
	}
}

type withDefaults struct {
	Name   string `codec:"name"`
	Weight uint16 `codec:"weight,default=100"`
	Mode   string `codec:"mode,default=auto"`
}

type nestedDefaults struct {
	Inner withDefaults `codec:"inner"`
}

func TestFromMap_Defaults(t *testing.T) {
	m := map[string]any{"name": "x"}
	var s withDefaults
	if err := FromMap(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "x" {
		t.Errorf("name = %q, want %q", s.Name, "x")
	}
	if s.Weight != 100 {
		t.Errorf("weight = %d, want 100", s.Weight)
	}
	if s.Mode != "auto" {
		t.Errorf("mode = %q, want %q", s.Mode, "auto")
	}
}

func TestFromMap_DefaultsOverridden(t *testing.T) {
	m := map[string]any{"name": "x", "weight": 50, "mode": "manual"}
	var s withDefaults
	if err := FromMap(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Weight != 50 {
		t.Errorf("weight = %d, want 50", s.Weight)
	}
	if s.Mode != "manual" {
		t.Errorf("mode = %q, want %q", s.Mode, "manual")
	}
}

func TestFromMap_DefaultsNested(t *testing.T) {
	m := map[string]any{"inner": map[string]any{"name": "n"}}
	var s nestedDefaults
	if err := FromMap(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Inner.Weight != 100 {
		t.Errorf("inner weight = %d, want 100", s.Inner.Weight)
	}
	if s.Inner.Mode != "auto" {
		t.Errorf("inner mode = %q, want %q", s.Inner.Mode, "auto")
	}
}

func TestFromMap_DefaultsExplicitZero(t *testing.T) {
	m := map[string]any{"name": "x", "weight": 0, "mode": ""}
	var s withDefaults
	if err := FromMap(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Weight != 0 {
		t.Errorf("weight = %d, want 0", s.Weight)
	}
	if s.Mode != "" {
		t.Errorf("mode = %q, want %q", s.Mode, "")
	}
}

type allDefaults struct {
	B bool    `codec:"b,default=true"`
	I int64   `codec:"i,default=-42"`
	F float64 `codec:"f,default=3.14"`
}

func TestFromMap_DefaultBool(t *testing.T) {
	var s allDefaults
	if err := FromMap(map[string]any{}, &s); err != nil {
		t.Fatal(err)
	}
	if s.B != true {
		t.Errorf("b = %v, want true", s.B)
	}
}

func TestFromMap_DefaultInt(t *testing.T) {
	var s allDefaults
	if err := FromMap(map[string]any{}, &s); err != nil {
		t.Fatal(err)
	}
	if s.I != -42 {
		t.Errorf("i = %d, want -42", s.I)
	}
}

func TestFromMap_DefaultFloat(t *testing.T) {
	var s allDefaults
	if err := FromMap(map[string]any{}, &s); err != nil {
		t.Fatal(err)
	}
	if s.F != 3.14 {
		t.Errorf("f = %f, want 3.14", s.F)
	}
}

func TestFromMap_DefaultUnsupportedType(t *testing.T) {
	type unsupported struct {
		Sl []int `codec:"sl,default=nope"`
	}
	var s unsupported
	err := FromMap(map[string]any{}, &s)
	if err == nil {
		t.Fatal("expected error for unsupported default type")
	}
}

type validatable struct {
	Name string `codec:"name"`
}

func (v *validatable) Validate() error {
	if v.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func TestValidate(t *testing.T) {
	v := &validatable{Name: "ok"}
	if err := Validate(v); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_Error(t *testing.T) {
	v := &validatable{}
	if err := Validate(v); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidate_NotValidatable(t *testing.T) {
	s := &sample{Name: "x"}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for non-Validatable type")
	}
}
