package cgroup

import (
	"errors"
	"testing"
)

func TestParseIOWeightDevice(t *testing.T) {
	got, err := parseIOWeightDevice("8:16 200")
	if err != nil {
		t.Fatal(err)
	}
	if got != (ioWeightDevice{Major: 8, Minor: 16, Weight: 200}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIOWeightDeviceRemovalRejected(t *testing.T) {
	for _, in := range []string{"8:16", "8:16 default"} {
		t.Run(in, func(t *testing.T) {
			_, err := parseIOWeightDevice(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestParseIOWeightDeviceMalformed(t *testing.T) {
	for _, in := range []string{"", "8 16 200", "8:16 abc", "8:16 200 trailing"} {
		t.Run(in, func(t *testing.T) {
			_, err := parseIOWeightDevice(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestIOWeightDeviceEqualByDevice(t *testing.T) {
	a := ioWeightDevice{Major: 8, Minor: 16, Weight: 100}
	if !a.equal(ioWeightDevice{Major: 8, Minor: 16, Weight: 999}) {
		t.Fatal("same device should be equal")
	}
}

func TestIOWeightDeviceCheckRejectsDivergence(t *testing.T) {
	a := ioWeightDevice{Major: 8, Minor: 16, Weight: 100}
	b := ioWeightDevice{Major: 8, Minor: 16, Weight: 200}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOWeightDeviceMergeNoOp(t *testing.T) {
	a := ioWeightDevice{Major: 8, Minor: 16, Weight: 100}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestIOWeightDeviceCheckDifferentDeviceNoConflict(t *testing.T) {
	a := ioWeightDevice{Major: 8, Minor: 16, Weight: 100}
	b := ioWeightDevice{Major: 8, Minor: 32, Weight: 999}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}
