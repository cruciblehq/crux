package fcap

import (
	"errors"
	"testing"
)

func TestNewBuilderStartsEmpty(t *testing.T) {
	if NewBuilder().State() != nil {
		t.Fatal("state != nil")
	}
}

func TestBuilderBuildAccumulatesFileCapabilityState(t *testing.T) {
	b := NewBuilder()
	if err := b.Build("/usr/bin/ping effective net_raw"); err != nil {
		t.Fatal(err)
	}

	s := b.State()
	assertEntry(t, s.Entries["/usr/bin/ping"], []string{"net_raw"}, nil, true)
}

func TestBuilderBuildErrorLeavesStateEmpty(t *testing.T) {
	b := NewBuilder()
	err := b.Build("/usr/bin/ping effective not_a_cap")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
	if b.State() != nil {
		t.Fatal("state != nil")
	}
}

func TestBuilderStateReturnsIsolatedSnapshot(t *testing.T) {
	b := NewBuilder()
	if err := b.Build("/usr/bin/ping effective net_raw"); err != nil {
		t.Fatal(err)
	}

	s := b.State()
	s.Entries["/usr/bin/ping"].Permitted[0] = "net_admin"
	s.Entries["/usr/bin/ping"].Effective = false

	again := b.State()
	assertEntry(t, again.Entries["/usr/bin/ping"], []string{"net_raw"}, nil, true)
}

func TestNewBuilderWithStateClonesInput(t *testing.T) {
	seed := &State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted: []string{"net_raw"},
			Effective: true,
		},
	}}
	b := NewBuilderWithState(seed)
	seed.Entries["/usr/bin/ping"].Permitted[0] = "net_admin"
	seed.Entries["/usr/bin/ping"].Effective = false

	s := b.State()
	assertEntry(t, s.Entries["/usr/bin/ping"], []string{"net_raw"}, nil, true)
}

func TestBuilderMergeNilIsNoOp(t *testing.T) {
	b := NewBuilder()
	if err := b.Merge(nil); err != nil {
		t.Fatal(err)
	}
	if b.State() != nil {
		t.Fatal("state != nil")
	}
}

func TestBuilderMergeClonesInputWhenEmpty(t *testing.T) {
	other := &State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted: []string{"net_raw"},
			Effective: true,
		},
	}}
	b := NewBuilder()
	if err := b.Merge(other); err != nil {
		t.Fatal(err)
	}
	other.Entries["/usr/bin/ping"].Permitted[0] = "net_admin"
	other.Entries["/usr/bin/ping"].Effective = false

	s := b.State()
	assertEntry(t, s.Entries["/usr/bin/ping"], []string{"net_raw"}, nil, true)
}

func TestBuilderMergeMergesEntriesIdempotently(t *testing.T) {
	b := NewBuilder()
	if err := b.Merge(&State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted: []string{"net_raw"},
			Effective: true,
		},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := b.Merge(&State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted:   []string{"net_raw"},
			Inheritable: []string{"sys_admin"},
		},
	}}); err != nil {
		t.Fatal(err)
	}

	s := b.State()
	assertEntry(t, s.Entries["/usr/bin/ping"], []string{"net_raw"}, []string{"sys_admin"}, true)
}

func TestBuilderBuildRejectsNonCleanPath(t *testing.T) {
	b := NewBuilder()
	if err := b.Build("/usr/bin/ping effective net_raw"); err != nil {
		t.Fatal(err)
	}
	err := b.Build("/usr//bin///ping inheritable sys_admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}

	s := b.State()
	if len(s.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(s.Entries))
	}
	assertEntry(t, s.Entries["/usr/bin/ping"], []string{"net_raw"}, nil, true)
}
