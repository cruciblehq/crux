package cgroup

import (
	"errors"
	"testing"
)

func TestParseNodeType(t *testing.T) {
	for _, in := range []string{"domain", "threaded"} {
		t.Run(in, func(t *testing.T) {
			got, err := parseNodeType(in)
			if err != nil || string(got) != in {
				t.Fatalf("got %q err %v", got, err)
			}
		})
	}
	for _, in := range []string{"", "Domain", "thread"} {
		t.Run("invalid_"+in, func(t *testing.T) {
			_, err := parseNodeType(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestMergeSubtreeControlFreshList(t *testing.T) {
	s := newSpec()
	added, err := s.mergeSubtreeControl([]controller{controllerCPU, controllerMemory})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !added {
		t.Fatal("added = false, want true")
	}
	if len(s.Cgroup.SubtreeControl) != 2 {
		t.Fatalf("SubtreeControl = %v", s.Cgroup.SubtreeControl)
	}
}

func TestMergeSubtreeControlEmptyIsNoOp(t *testing.T) {
	s := newSpec()
	added, err := s.mergeSubtreeControl(nil)
	if err != nil || added {
		t.Fatalf("got (%v, %v)", added, err)
	}
}

func TestMergeSubtreeControlSameListIdempotent(t *testing.T) {
	s := newSpec()
	if _, err := s.mergeSubtreeControl([]controller{controllerCPU}); err != nil {
		t.Fatal(err)
	}
	added, err := s.mergeSubtreeControl([]controller{controllerCPU})
	if err != nil || added {
		t.Fatalf("got (%v, %v)", added, err)
	}
}

func TestMergeSubtreeControlDifferentListConflicts(t *testing.T) {
	s := newSpec()
	if _, err := s.mergeSubtreeControl([]controller{controllerCPU}); err != nil {
		t.Fatal(err)
	}
	_, err := s.mergeSubtreeControl([]controller{controllerMemory})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}
