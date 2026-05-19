package cgroup

import (
	"errors"
	"slices"
	"testing"
)

func TestParseControllerValid(t *testing.T) {
	for _, in := range []controller{
		controllerCPU, controllerCPUSet, controllerIO, controllerMemory,
		controllerHugeTLB, controllerPids, controllerRDMA, controllerMisc,
		controllerDevMem,
	} {
		t.Run(string(in), func(t *testing.T) {
			got, err := parseController(string(in))
			if err != nil || got != in {
				t.Fatalf("got %q err %v", got, err)
			}
		})
	}
}

func TestParseControllerInvalid(t *testing.T) {
	for _, in := range []string{"", "bogus", "CPU"} {
		t.Run(in, func(t *testing.T) {
			_, err := parseController(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestParseSubtreeControl(t *testing.T) {
	got, err := parseSubtreeControl("cpu memory io")
	if err != nil {
		t.Fatal(err)
	}
	want := []controller{controllerCPU, controllerMemory, controllerIO}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseSubtreeControlEmpty(t *testing.T) {
	_, err := parseSubtreeControl("")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseSubtreeControlRejectsPrefix(t *testing.T) {
	for _, in := range []string{"+cpu", "-cpu", "cpu +memory"} {
		t.Run(in, func(t *testing.T) {
			_, err := parseSubtreeControl(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestParseSubtreeControlRejectsDuplicates(t *testing.T) {
	_, err := parseSubtreeControl("cpu cpu")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseSubtreeControlRejectsUnknown(t *testing.T) {
	_, err := parseSubtreeControl("cpu bogus")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}
