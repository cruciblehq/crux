package rlimit

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/resource/affordance/agl"
)

func nameArg(v string) agl.Arg { return agl.Arg{Type: agl.ArgName, Value: v} }
func intArg(v string) agl.Arg  { return agl.Arg{Type: agl.ArgInt, Value: v} }

// Returns a Subsystem with an empty rlimits slice.
func newSub() (*Subsystem, *[]specs.POSIXRlimit) {
	rl := []specs.POSIXRlimit{}
	return New(&rl), &rl
}

func TestBuildAppliesGrant(t *testing.T) {
	sub, rl := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{nameArg("nofile"), intArg("1024"), intArg("4096")},
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
	if err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{nameArg("core"), intArg("0")},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if (*rl)[0].Hard != 0 {
		t.Fatalf("hard = %d, want 0", (*rl)[0].Hard)
	}
}

func TestBuildSoftExceedsHard(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{nameArg("nofile"), intArg("4096"), intArg("1024")},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args:  []agl.Arg{nameArg("nofile"), intArg("1024")},
		Where: &agl.CompareExpr{},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args:   []agl.Arg{nameArg("nofile"), intArg("1024")},
		Kwargs: []agl.Kwarg{{Key: "k", Value: nameArg("v")}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownResource(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{nameArg("bogus"), intArg("1")},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

