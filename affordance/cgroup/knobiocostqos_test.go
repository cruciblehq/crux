package cgroup

import (
	"errors"
	"testing"
)

func TestParseIOCostQoS(t *testing.T) {
	got, err := parseIOCostQoS("8 0 enable=true ctrl=user rpct=95 rlat=5000 wpct=95 wlat=5000 min=50 max=150")
	if err != nil {
		t.Fatal(err)
	}
	want := ioCostQoS{Major: 8, Minor: 0, Enable: true, Ctrl: ioCtrlModeUser, Rpct: 95, Rlat: 5000, Wpct: 95, Wlat: 5000, Min: 50, Max: 150}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseIOCostQoSBareIdentity(t *testing.T) {
	got, err := parseIOCostQoS("8 0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Major != 8 || got.Minor != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIOCostQoSInvalidEnumValue(t *testing.T) {
	_, err := parseIOCostQoS("8 0 ctrl=bogus")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOCostQoSEqualByDevice(t *testing.T) {
	a := ioCostQoS{Major: 8, Minor: 0, Rpct: 1}
	if !a.equal(ioCostQoS{Major: 8, Minor: 0, Rpct: 99}) {
		t.Fatal("same device should be equal")
	}
	if a.equal(ioCostQoS{Major: 8, Minor: 16}) {
		t.Fatal("different minor should not be equal")
	}
}

func TestIOCostQoSCheckRejectsDivergence(t *testing.T) {
	a := ioCostQoS{Major: 8, Minor: 0, Rpct: 50}
	b := ioCostQoS{Major: 8, Minor: 0, Rpct: 60}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOCostQoSMergeNoOp(t *testing.T) {
	a := ioCostQoS{Major: 8, Minor: 0, Rpct: 50}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestIOCostQoSCheckDifferentDeviceNoConflict(t *testing.T) {
	a := ioCostQoS{Major: 8, Minor: 0, Rpct: 50}
	b := ioCostQoS{Major: 8, Minor: 16, Rpct: 99}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}
