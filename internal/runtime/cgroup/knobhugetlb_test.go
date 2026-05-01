package cgroup

import (
	"errors"
	"testing"
)

func TestParseHugeTLBSizeOnly(t *testing.T) {
	got, err := parseHugeTLB("2MB")
	if err != nil {
		t.Fatal(err)
	}
	if got.Size != "2MB" || got.Max != 0 || got.RsvdMax != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseHugeTLBWithKeys(t *testing.T) {
	got, err := parseHugeTLB("1GB max=2 rsvd_max=1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Size != "1GB" || got.Max != 2 || got.RsvdMax != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseHugeTLBRequiresSize(t *testing.T) {
	_, err := parseHugeTLB("")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseHugeTLBUnknownKey(t *testing.T) {
	_, err := parseHugeTLB("2MB foo=1")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestHugeTLBEqualBySize(t *testing.T) {
	a := hugeTLB{Size: "2MB", Max: 1}
	b := hugeTLB{Size: "2MB", Max: 99}
	if !a.equal(b) {
		t.Fatal("same-size entries should compare equal")
	}
	if a.equal(hugeTLB{Size: "1GB"}) {
		t.Fatal("different size should not be equal")
	}
}

func TestHugeTLBCheckRejectsDivergence(t *testing.T) {
	a := hugeTLB{Size: "2MB", Max: 1}
	b := hugeTLB{Size: "2MB", Max: 2}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestHugeTLBCheckIdenticalNoConflict(t *testing.T) {
	a := hugeTLB{Size: "2MB", Max: 1, RsvdMax: 0}
	if err := a.check(a); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestHugeTLBMergeNoOp(t *testing.T) {
	a := hugeTLB{Size: "2MB", Max: 1}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestHugeTLBCheckDifferentSizeNoConflict(t *testing.T) {
	a := hugeTLB{Size: "2MB", Max: 1}
	b := hugeTLB{Size: "1GB", Max: 999}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}
