package cgroup

import (
	"errors"
	"testing"
)

func TestParseIOMax(t *testing.T) {
	got, err := parseIOMax("8 0 rbps=1000 wbps=2000 riops=10 wiops=20")
	if err != nil {
		t.Fatal(err)
	}
	want := ioMax{Major: 8, Minor: 0, Rbps: 1000, Wbps: 2000, Riops: 10, Wiops: 20}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseIOMaxBareIdentity(t *testing.T) {
	got, err := parseIOMax("8 0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Major != 8 || got.Minor != 0 || got.Rbps != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIOMaxUnknownKey(t *testing.T) {
	_, err := parseIOMax("8 0 bogus=1")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOMaxEqualByDevice(t *testing.T) {
	a := ioMax{Major: 8, Minor: 0, Rbps: 1}
	if !a.equal(ioMax{Major: 8, Minor: 0, Rbps: 99}) {
		t.Fatal("same device should be equal")
	}
}

func TestIOMaxCheckRejectsDivergence(t *testing.T) {
	a := ioMax{Major: 8, Minor: 0, Rbps: 1}
	b := ioMax{Major: 8, Minor: 0, Rbps: 2}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOMaxMergeNoOp(t *testing.T) {
	a := ioMax{Major: 8, Minor: 0, Rbps: 1}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestIOMaxCheckDifferentDeviceNoConflict(t *testing.T) {
	a := ioMax{Major: 8, Minor: 0, Rbps: 1}
	b := ioMax{Major: 8, Minor: 16, Rbps: 999}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}
