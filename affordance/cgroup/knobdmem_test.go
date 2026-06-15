package cgroup

import (
	"errors"
	"testing"
)

func TestParseDmemValueWithValue(t *testing.T) {
	got, err := parseDmemValue("gpu0 1024", func(e *dmem, v uint64) { e.Max = v })
	if err != nil {
		t.Fatal(err)
	}
	if got.Region != "gpu0" || got.Max != 1024 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseDmemValueWithoutValue(t *testing.T) {
	got, err := parseDmemValue("gpu0", func(e *dmem, v uint64) { e.Max = v })
	if err != nil {
		t.Fatal(err)
	}
	if got.Region != "gpu0" || got.Max != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseDmemValueRequiresRegion(t *testing.T) {
	_, err := parseDmemValue("", func(*dmem, uint64) {})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseDmemValueInvalidValue(t *testing.T) {
	_, err := parseDmemValue("gpu0 nope", func(e *dmem, v uint64) { e.Max = v })
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseDmemEntryRoutesByKnob(t *testing.T) {
	cases := []struct {
		knob  string
		check func(dmem) bool
	}{
		{dmemMaxKnob, func(d dmem) bool { return d.Max == 100 }},
		{dmemMinKnob, func(d dmem) bool { return d.Min == 100 }},
		{dmemLowKnob, func(d dmem) bool { return d.Low == 100 }},
	}
	for _, c := range cases {
		t.Run(c.knob, func(t *testing.T) {
			got, err := parseDmemEntry(c.knob, "gpu0 100")
			if err != nil {
				t.Fatal(err)
			}
			if !c.check(got) {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func TestParseDmemEntryUnknownKnob(t *testing.T) {
	_, err := parseDmemEntry("dmem.bogus", "gpu0 1")
	if !errors.Is(err, ErrUnknownKnob) {
		t.Fatalf("err = %v", err)
	}
}

func TestDmemEqualByRegion(t *testing.T) {
	a := dmem{Region: "gpu0", Max: 1}
	b := dmem{Region: "gpu0", Min: 2}
	if !a.equal(b) {
		t.Fatal("same-region entries should compare equal")
	}
	if a.equal(dmem{Region: "gpu1"}) {
		t.Fatal("different region should not be equal")
	}
}

func TestDmemCheckPerFieldConflict(t *testing.T) {
	a := dmem{Region: "gpu0", Max: 1, Min: 2, Low: 3}
	for _, b := range []dmem{
		{Region: "gpu0", Max: 99},
		{Region: "gpu0", Min: 99},
		{Region: "gpu0", Low: 99},
	} {
		if err := a.check(b); !errors.Is(err, ErrConflict) {
			t.Fatalf("expected conflict for %+v, got %v", b, err)
		}
	}
}

func TestDmemCheckDifferentRegionNoConflict(t *testing.T) {
	a := dmem{Region: "gpu0", Max: 1}
	b := dmem{Region: "gpu1", Max: 99}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestDmemMergeUnion(t *testing.T) {
	a := dmem{Region: "gpu0", Max: 10}
	b := dmem{Region: "gpu0", Min: 5, Low: 15}
	if !a.merge(b) {
		t.Fatal("merge added fields but reported no change")
	}
	if a.Max != 10 || a.Min != 5 || a.Low != 15 {
		t.Fatalf("a = %+v", a)
	}
}

func TestDmemMergeNoChange(t *testing.T) {
	a := dmem{Region: "gpu0", Max: 10, Min: 5, Low: 15}
	if a.merge(a) {
		t.Fatal("merge of identical entries reported change")
	}
}
