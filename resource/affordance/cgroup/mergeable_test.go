package cgroup

import (
	"errors"
	"testing"
)

// Single-key entry used to exercise the merge generic in isolation.
type kv struct {
	K string
	V int
}

func (e kv) equal(o kv) bool {
	return e.K == o.K
}

func (e kv) check(o kv) error {
	if e.K == o.K && e.V != o.V {
		return ErrConflict
	}
	return nil
}
func (e *kv) merge(o kv) bool {
	if e.V == o.V {
		return false
	}
	e.V = o.V
	return true
}

func TestMergeAppendsNewIdentity(t *testing.T) {
	var s []kv
	changed, err := merge(&s, kv{"a", 1})
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	if len(s) != 1 || s[0] != (kv{"a", 1}) {
		t.Fatalf("s = %v", s)
	}
}

func TestMergeIdempotentForSameValue(t *testing.T) {
	s := []kv{{"a", 1}}
	changed, err := merge(&s, kv{"a", 1})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if changed {
		t.Fatal("changed = true, want false for identical entry")
	}
}

func TestMergeDetectsConflict(t *testing.T) {
	s := []kv{{"a", 1}}
	_, err := merge(&s, kv{"a", 2})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestMergeAppendsDifferentIdentities(t *testing.T) {
	var s []kv
	for _, e := range []kv{{"a", 1}, {"b", 2}, {"c", 3}} {
		if _, err := merge(&s, e); err != nil {
			t.Fatalf("err = %v", err)
		}
	}
	if len(s) != 3 {
		t.Fatalf("len = %d, want 3", len(s))
	}
}
