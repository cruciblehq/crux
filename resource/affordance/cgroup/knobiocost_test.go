package cgroup

import (
	"errors"
	"testing"
)

func TestParseIOCost(t *testing.T) {
	got, err := parseIOCost("8 0 rbps=1000 wbps=2000 rseqiops=10 rrandiops=5 wseqiops=20 wrandiops=8")
	if err != nil {
		t.Fatal(err)
	}
	want := ioCost{Major: 8, Minor: 0, Rbps: 1000, Wbps: 2000, Rseqiops: 10, Rrandiops: 5, Wseqiops: 20, Wrandiops: 8}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseIOCostBareIdentity(t *testing.T) {
	got, err := parseIOCost("8 0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Major != 8 || got.Minor != 0 || got.Rbps != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIOCostInvalidKey(t *testing.T) {
	_, err := parseIOCost("8 0 foo=1")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOCostEqualByDevice(t *testing.T) {
	a := ioCost{Major: 8, Minor: 0, Rbps: 1}
	if !a.equal(ioCost{Major: 8, Minor: 0, Rbps: 99}) {
		t.Fatal("same device should be equal")
	}
	if a.equal(ioCost{Major: 8, Minor: 16}) {
		t.Fatal("different minor should not be equal")
	}
}

func TestIOCostCheckRejectsDivergence(t *testing.T) {
	a := ioCost{Major: 8, Minor: 0, Rbps: 1}
	b := ioCost{Major: 8, Minor: 0, Rbps: 2}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOCostCheckIdenticalNoConflict(t *testing.T) {
	a := ioCost{Major: 8, Minor: 0, Rbps: 1}
	if err := a.check(a); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestIOCostCheckDifferentDeviceNoConflict(t *testing.T) {
	a := ioCost{Major: 8, Minor: 0, Rbps: 1}
	b := ioCost{Major: 8, Minor: 16, Rbps: 999}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestIOCostMergeNoOp(t *testing.T) {
	a := ioCost{Major: 8, Minor: 0, Rbps: 1}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}
