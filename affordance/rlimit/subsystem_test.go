package rlimit

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
)

func nameArg(v string) agl.Arg {
	return agl.Arg{Type: agl.ArgName, Value: v}
}

func intArg(v string) agl.Arg {
	return agl.Arg{Type: agl.ArgInt, Value: v}
}

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

func TestNameReturnsRlimit(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != "rlimit" {
		t.Fatalf("Name() = %q, want %q", got, "rlimit")
	}
}

func TestKeyReturnsResourceName(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("nofile"), intArg("1024")}}
	if got := sub.Key(&g); got != "nofile" {
		t.Fatalf("Key() = %q, want %q", got, "nofile")
	}
}

func TestKeyEmptyWhenNoArgs(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Key(&agl.Model{}); got != "" {
		t.Fatalf("Key() = %q, want empty", got)
	}
}

func TestApplyUpdatesInPlace(t *testing.T) {
	rl := []specs.POSIXRlimit{{Type: "RLIMIT_NOFILE", Soft: 0, Hard: 0}}
	sub := New(&rl)
	if err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{nameArg("nofile"), intArg("1024"), intArg("4096")},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(rl) != 1 {
		t.Fatalf("len(rl) = %d, want 1 (should update in-place)", len(rl))
	}
	if rl[0].Soft != 1024 || rl[0].Hard != 4096 {
		t.Fatalf("rl[0] = %+v, want Soft=1024 Hard=4096", rl[0])
	}
}

func TestParseLimitUnlimited(t *testing.T) {
	v, err := parseLimit(nameArg("unlimited"), "soft")
	if err != nil {
		t.Fatalf("parseLimit unlimited: %v", err)
	}
	if v != ^uint64(0) {
		t.Fatalf("v = %d, want MaxUint64", v)
	}
}

func TestParseLimitNonIntNonName(t *testing.T) {
	_, err := parseLimit(agl.Arg{Type: agl.ArgStrASCII, Value: "abc"}, "soft")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonNameResource(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{intArg("42"), intArg("1")},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildHardLimitParseError(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{nameArg("nofile"), intArg("1024"), {Type: agl.ArgStrASCII, Value: "x"}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWrongArgCount(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "rlimit", Args: []agl.Arg{nameArg("nofile")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildSoftLimitParseError(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Subsystem: "rlimit",
		Args:      []agl.Arg{nameArg("nofile"), {Type: agl.ArgStrASCII, Value: "x"}},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestApplySkipsNonMatchingEntry(t *testing.T) {
	rl := []specs.POSIXRlimit{
		{Type: "RLIMIT_CORE", Soft: 0, Hard: 0},
		{Type: "RLIMIT_NOFILE", Soft: 0, Hard: 0},
	}
	sub := New(&rl)
	if err := sub.Build(&agl.Model{Subsystem: "rlimit",
		Args: []agl.Arg{nameArg("nofile"), intArg("1024"), intArg("4096")},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(rl) != 2 {
		t.Fatalf("len(rl) = %d, want 2 (should update in-place)", len(rl))
	}
	if rl[1].Soft != 1024 || rl[1].Hard != 4096 {
		t.Fatalf("rl[1] = %+v, want Soft=1024 Hard=4096", rl[1])
	}
}

func TestParseLimitOverflow(t *testing.T) {
	_, err := parseLimit(agl.Arg{Type: agl.ArgInt, Value: "18446744073709551616"}, "soft")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
