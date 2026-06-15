package cgroup

import (
	"errors"
	"testing"
)

func TestParseIOLatency(t *testing.T) {
	got, err := parseIOLatency("8 0 target=2000")
	if err != nil {
		t.Fatal(err)
	}
	if got != (ioLatency{Major: 8, Minor: 0, Target: 2000}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIOLatencyBareIdentity(t *testing.T) {
	got, err := parseIOLatency("8 0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Major != 8 || got.Minor != 0 || got.Target != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIOLatencyInvalidKey(t *testing.T) {
	_, err := parseIOLatency("8 0 latency=1")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOLatencyEqualByDevice(t *testing.T) {
	a := ioLatency{Major: 8, Minor: 0, Target: 1}
	if !a.equal(ioLatency{Major: 8, Minor: 0, Target: 99}) {
		t.Fatal("same device should be equal")
	}
}

func TestIOLatencyCheckRejectsDivergence(t *testing.T) {
	a := ioLatency{Major: 8, Minor: 0, Target: 1}
	b := ioLatency{Major: 8, Minor: 0, Target: 2}
	if err := a.check(b); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestIOLatencyMergeNoOp(t *testing.T) {
	a := ioLatency{Major: 8, Minor: 0, Target: 1}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestIOLatencyCheckDifferentDeviceNoConflict(t *testing.T) {
	a := ioLatency{Major: 8, Minor: 0, Target: 1}
	b := ioLatency{Major: 8, Minor: 16, Target: 99}
	if err := a.check(b); err != nil {
		t.Fatalf("err = %v", err)
	}
}
