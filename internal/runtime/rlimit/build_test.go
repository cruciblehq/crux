package rlimit

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/aegis"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

func nameArg(v string) aegis.Arg { return aegis.Arg{Type: aegis.ArgName, Value: v} }
func intArg(v string) aegis.Arg  { return aegis.Arg{Type: aegis.ArgInt, Value: v} }

// Returns a Subsystem with an empty rlimits slice.
func newSub() (*Subsystem, *[]specs.POSIXRlimit) {
	rl := []specs.POSIXRlimit{}
	return New(&rl), &rl
}

// Wraps a rlimits slice header into a unified spec for Merge inputs.
func wrap(rl *[]specs.POSIXRlimit) shared.Spec {
	return shared.Spec{OCI: &specs.Spec{Process: &specs.Process{Rlimits: *rl}}}
}

func TestBuildAppliesGrant(t *testing.T) {
	sub, rl := newSub()
	if err := sub.Build(&aegis.Model{Subsystem: "rlimit",
		Args: []aegis.Arg{nameArg("nofile"), intArg("1024"), intArg("4096")},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(*rl) != 1 {
		t.Fatalf("rlimits len = %d, want 1", len(*rl))
	}
	if l := (*rl)[0]; l.Type != "RLIMIT_NOFILE" || l.Soft != 1024 || l.Hard != 4096 {
		t.Fatalf("got %+v", l)
	}
}

func TestBuildHardDefaultsToSoft(t *testing.T) {
	sub, rl := newSub()
	if err := sub.Build(&aegis.Model{Subsystem: "rlimit",
		Args: []aegis.Arg{nameArg("core"), intArg("0")},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if (*rl)[0].Hard != 0 {
		t.Fatalf("hard = %d, want 0", (*rl)[0].Hard)
	}
}

func TestBuildIdempotent(t *testing.T) {
	sub, rl := newSub()
	g := aegis.Model{Subsystem: "rlimit", Args: []aegis.Arg{nameArg("nofile"), intArg("1024")}}
	if err := sub.Build(&g); err != nil {
		t.Fatal(err)
	}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("second Build: %v", err)
	}
	if len(*rl) != 1 {
		t.Fatalf("rlimits len = %d, want 1", len(*rl))
	}
}

func TestBuildConflictRejected(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&aegis.Model{Subsystem: "rlimit", Args: []aegis.Arg{nameArg("nofile"), intArg("1024")}}); err != nil {
		t.Fatal(err)
	}
	err := sub.Build(&aegis.Model{Subsystem: "rlimit", Args: []aegis.Arg{nameArg("nofile"), intArg("2048")}})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestBuildSoftExceedsHard(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&aegis.Model{Subsystem: "rlimit",
		Args: []aegis.Arg{nameArg("nofile"), intArg("4096"), intArg("1024")},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&aegis.Model{Subsystem: "rlimit",
		Args:  []aegis.Arg{nameArg("nofile"), intArg("1024")},
		Where: &aegis.CompareExpr{},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&aegis.Model{Subsystem: "rlimit",
		Args:   []aegis.Arg{nameArg("nofile"), intArg("1024")},
		Kwargs: []aegis.Kwarg{{Key: "k", Value: nameArg("v")}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownResource(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&aegis.Model{Subsystem: "rlimit",
		Args: []aegis.Arg{nameArg("bogus"), intArg("1")},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestMergeUnionsAndConflicts(t *testing.T) {
	sub, dst := newSub()
	if err := apply(dst, specs.POSIXRlimit{Type: "RLIMIT_NOFILE", Soft: 1024, Hard: 1024}); err != nil {
		t.Fatal(err)
	}
	src := &[]specs.POSIXRlimit{
		{Type: "RLIMIT_CORE", Soft: 0, Hard: 0},
		{Type: "RLIMIT_NOFILE", Soft: 1024, Hard: 1024},
	}
	if err := sub.Merge(wrap(src)); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(*dst) != 2 {
		t.Fatalf("rlimits len = %d, want 2", len(*dst))
	}
	conflict := &[]specs.POSIXRlimit{{Type: "RLIMIT_NOFILE", Soft: 2048, Hard: 2048}}
	if err := sub.Merge(wrap(conflict)); !errors.Is(err, ErrConflict) {
		t.Fatalf("Merge conflict err = %v, want ErrConflict", err)
	}
}

func TestMergeNilIsNoOp(t *testing.T) {
	sub, rl := newSub()
	if err := apply(rl, specs.POSIXRlimit{Type: "RLIMIT_NOFILE", Soft: 1024, Hard: 1024}); err != nil {
		t.Fatal(err)
	}
	if err := sub.Merge(shared.Spec{}); err != nil {
		t.Fatalf("Merge(empty): %v", err)
	}
	if len(*rl) != 1 {
		t.Fatal("Merge(empty) mutated state")
	}
}
