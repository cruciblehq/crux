package codec

import "testing"

type withDefaults struct {
	Name   string `codec:"name"`
	Weight uint16 `codec:"weight,default=100"`
	Mode   string `codec:"mode,default=auto"`
}

type nestedDefaults struct {
	Inner withDefaults `codec:"inner"`
}

type allDefaults struct {
	B bool    `codec:"b,default=true"`
	I int64   `codec:"i,default=-42"`
	F float64 `codec:"f,default=3.14"`
}

func TestDecode_Defaults(t *testing.T) {
	m := map[string]any{"name": "x"}
	var s withDefaults
	if err := Decode(m, &s); err != nil {
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

func TestDecode_DefaultsOverridden(t *testing.T) {
	m := map[string]any{"name": "x", "weight": 50, "mode": "manual"}
	var s withDefaults
	if err := Decode(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Weight != 50 {
		t.Errorf("weight = %d, want 50", s.Weight)
	}
	if s.Mode != "manual" {
		t.Errorf("mode = %q, want %q", s.Mode, "manual")
	}
}

func TestDecode_DefaultsNested(t *testing.T) {
	m := map[string]any{"inner": map[string]any{"name": "n"}}
	var s nestedDefaults
	if err := Decode(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Inner.Weight != 100 {
		t.Errorf("inner weight = %d, want 100", s.Inner.Weight)
	}
	if s.Inner.Mode != "auto" {
		t.Errorf("inner mode = %q, want %q", s.Inner.Mode, "auto")
	}
}

func TestDecode_DefaultsExplicitZero(t *testing.T) {
	m := map[string]any{"name": "x", "weight": 0, "mode": ""}
	var s withDefaults
	if err := Decode(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Weight != 0 {
		t.Errorf("weight = %d, want 0", s.Weight)
	}
	if s.Mode != "" {
		t.Errorf("mode = %q, want %q", s.Mode, "")
	}
}

func TestDecode_DefaultBool(t *testing.T) {
	var s allDefaults
	if err := Decode(map[string]any{}, &s); err != nil {
		t.Fatal(err)
	}
	if s.B != true {
		t.Errorf("b = %v, want true", s.B)
	}
}

func TestDecode_DefaultInt(t *testing.T) {
	var s allDefaults
	if err := Decode(map[string]any{}, &s); err != nil {
		t.Fatal(err)
	}
	if s.I != -42 {
		t.Errorf("i = %d, want -42", s.I)
	}
}

func TestDecode_DefaultFloat(t *testing.T) {
	var s allDefaults
	if err := Decode(map[string]any{}, &s); err != nil {
		t.Fatal(err)
	}
	if s.F != 3.14 {
		t.Errorf("f = %f, want 3.14", s.F)
	}
}

func TestDecode_DefaultUnsupportedType(t *testing.T) {
	type unsupported struct {
		Sl []int `codec:"sl,default=nope"`
	}
	var s unsupported
	err := Decode(map[string]any{}, &s)
	if err == nil {
		t.Fatal("expected error for unsupported default type")
	}
}
