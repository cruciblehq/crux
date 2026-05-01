package cgroup

import (
	"errors"
	"testing"
)

func TestParseIOPrioClass(t *testing.T) {
	for _, in := range []ioPrioClass{ioPrioClassRT, ioPrioClassBE, ioPrioClassIdle} {
		got, err := parseIOPrioClass(string(in))
		if err != nil || got != in {
			t.Fatalf("parseIOPrioClass(%q) = %q, %v", in, got, err)
		}
	}
	for _, in := range []string{"", "RT", "high"} {
		_, err := parseIOPrioClass(in)
		if !errors.Is(err, ErrInvalidGrant) {
			t.Fatalf("%q err = %v", in, err)
		}
	}
}

func TestParseIOCtrlMode(t *testing.T) {
	for _, in := range []ioCtrlMode{ioCtrlModeAuto, ioCtrlModeUser} {
		got, err := parseIOCtrlMode(string(in))
		if err != nil || got != in {
			t.Fatalf("parseIOCtrlMode(%q) = %q, %v", in, got, err)
		}
	}
	for _, in := range []string{"", "AUTO", "manual"} {
		_, err := parseIOCtrlMode(in)
		if !errors.Is(err, ErrInvalidGrant) {
			t.Fatalf("%q err = %v", in, err)
		}
	}
}

func TestParseIOWeightScalarPlain(t *testing.T) {
	got, err := parseIOWeightScalar([]string{"100"})
	if err != nil || got != 100 {
		t.Fatalf("got %d err %v", got, err)
	}
}

func TestParseIOWeightScalarDefaultPrefix(t *testing.T) {
	got, err := parseIOWeightScalar([]string{"default", "250"})
	if err != nil || got != 250 {
		t.Fatalf("got %d err %v", got, err)
	}
}

func TestParseIOWeightScalarDefaultMissingValue(t *testing.T) {
	_, err := parseIOWeightScalar([]string{"default"})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseIOWeightScalarOverflow(t *testing.T) {
	_, err := parseIOWeightScalar([]string{"65536"})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}
