package fcap

import (
	"errors"
	"testing"
)

func TestFcapValidateEmptyPath(t *testing.T) {
	f := &Spec{Entries: map[string]*Capabilities{
		"": {Permitted: []string{"CAP_CHOWN"}},
	}}
	if err := f.Validate(); !errors.Is(err, ErrInvalidFcap) {
		t.Fatalf("err = %v, want ErrInvalidFcap", err)
	}
}

func TestFcapValidateNilCaps(t *testing.T) {
	f := &Spec{Entries: map[string]*Capabilities{
		"/bin/prog": nil,
	}}
	if err := f.Validate(); !errors.Is(err, ErrInvalidFcap) {
		t.Fatalf("err = %v, want ErrInvalidFcap", err)
	}
}

func TestFcapValidateInvalidCaps(t *testing.T) {
	f := &Spec{Entries: map[string]*Capabilities{
		"/bin/prog": {},
	}}
	if err := f.Validate(); !errors.Is(err, ErrInvalidFcap) {
		t.Fatalf("err = %v, want ErrInvalidFcap", err)
	}
}

func TestFcapValidateOK(t *testing.T) {
	f := &Spec{Entries: map[string]*Capabilities{
		"/bin/prog": {Permitted: []string{"CAP_CHOWN"}},
	}}
	if err := f.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
