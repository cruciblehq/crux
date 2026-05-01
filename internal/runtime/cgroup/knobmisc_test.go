package cgroup

import (
	"errors"
	"testing"
)

func TestParseMisc(t *testing.T) {
	got, err := parseMisc("sev max=10")
	if err != nil {
		t.Fatal(err)
	}
	if got != (misc{Resource: "sev", Max: 10}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParseMiscWithoutMax(t *testing.T) {
	got, err := parseMisc("sev")
	if err != nil {
		t.Fatal(err)
	}
	if got.Resource != "sev" || got.Max != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseMiscRequiresResource(t *testing.T) {
	_, err := parseMisc("")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseMiscUnknownKey(t *testing.T) {
	_, err := parseMisc("sev bogus=1")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestMiscEqualByResource(t *testing.T) {
	a := misc{Resource: "sev", Max: 1}
	if !a.equal(misc{Resource: "sev", Max: 99}) {
		t.Fatal("same resource should be equal")
	}
	if a.equal(misc{Resource: "sev_es"}) {
		t.Fatal("different resource should not be equal")
	}
}

func TestMiscCheckRejectsDivergence(t *testing.T) {
	a := misc{Resource: "sev", Max: 1}
	b := misc{Resource: "sev", Max: 2}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestMiscMergeNoOp(t *testing.T) {
	a := misc{Resource: "sev", Max: 1}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestMiscCheckDifferentResourceNoConflict(t *testing.T) {
	a := misc{Resource: "sev", Max: 1}
	b := misc{Resource: "sev_es", Max: 999}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}
