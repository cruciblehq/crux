package manifest

import (
	"errors"
	"testing"
)

func TestParamValidateOK(t *testing.T) {
	cases := []Param{
		{Name: "name"},
		{Name: "host_url"},
		{Name: "x", Default: "v"},
		{Name: "x", Default: 42},
		{Name: "x", Default: 3.14},
		{Name: "x", Default: true},
	}
	for _, p := range cases {
		if err := p.Validate(); err != nil {
			t.Errorf("Param{%+v}.Validate(): %v", p, err)
		}
	}
}

func TestParamValidateMissingName(t *testing.T) {
	err := (&Param{}).Validate()
	if !errors.Is(err, ErrMissingParamName) {
		t.Fatalf("err = %v, want ErrMissingParamName", err)
	}
}

func TestParamValidateInvalidName(t *testing.T) {
	for _, n := range []string{"Foo", "1foo", "_x", "foo-bar", "foo.bar", ""} {
		p := &Param{Name: n}
		if err := p.Validate(); err == nil {
			t.Errorf("name %q: expected error", n)
		}
	}
}

func TestParamValidateInvalidDefault(t *testing.T) {
	p := &Param{Name: "x", Default: []string{"a"}}
	err := p.Validate()
	if !errors.Is(err, ErrInvalidParamDefault) {
		t.Fatalf("err = %v, want ErrInvalidParamDefault", err)
	}
}
