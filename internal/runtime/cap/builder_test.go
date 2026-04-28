package cap

import (
	"errors"
	"slices"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
)

func TestComposeNilBuilderIsNoOp(t *testing.T) {
	b := NewBuilder()
	target := &specs.LinuxCapabilities{}
	if err := b.Compose(target); err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if len(target.Effective) != 0 {
		t.Fatal("nil builder produced capabilities")
	}
}

func TestComposeFullMode(t *testing.T) {
	b := NewBuilder()
	b.apply("net_admin", ModeFull)
	target := &specs.LinuxCapabilities{}
	if err := b.Compose(target); err != nil {
		t.Fatalf("Compose: %v", err)
	}
	want := []string{"CAP_NET_ADMIN"}
	for name, got := range map[string][]string{
		"effective":   target.Effective,
		"permitted":   target.Permitted,
		"inheritable": target.Inheritable,
		"bounding":    target.Bounding,
		"ambient":     target.Ambient,
	} {
		if !slices.Equal(got, want) {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
}

func TestComposeDedupesAgainstBaseline(t *testing.T) {
	b := NewBuilder()
	b.apply("chown", ModeEffective)
	target := &specs.LinuxCapabilities{Effective: []string{"CAP_CHOWN"}}
	if err := b.Compose(target); err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if !slices.Equal(target.Effective, []string{"CAP_CHOWN"}) {
		t.Fatalf("Effective = %v, want [CAP_CHOWN]", target.Effective)
	}
}

func TestSpecReturnsEmptyWhenNoGrants(t *testing.T) {
	b := NewBuilder()
	got := b.Spec()
	if got == nil {
		t.Fatal("Model() = nil, want empty value")
	}
	if len(got.Effective) != 0 || len(got.Permitted) != 0 ||
		len(got.Inheritable) != 0 || len(got.Bounding) != 0 || len(got.Ambient) != 0 {
		t.Fatalf("Model() = %+v, want all empty", got)
	}
}

func TestSpecClonesAccumulatedState(t *testing.T) {
	b := NewBuilder()
	b.apply("net_admin", ModeBound)
	got := b.Spec()
	if got == nil {
		t.Fatal("Model() returned nil")
	}
	if !slices.Equal(got.Bounding, []string{"CAP_NET_ADMIN"}) {
		t.Fatalf("Bounding = %v", got.Bounding)
	}
	got.Bounding[0] = "CAP_MUTATED"
	if b.caps.Bounding[0] != "CAP_NET_ADMIN" {
		t.Fatal("Model() did not clone")
	}
}

func TestMergeUnionsSets(t *testing.T) {
	b := NewBuilder()
	b.apply("chown", ModeBound)
	other := &specs.LinuxCapabilities{
		Bounding:  []string{"CAP_CHOWN", "CAP_NET_ADMIN"},
		Effective: []string{"CAP_KILL"},
	}
	if err := b.Merge(other); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !slices.Equal(b.caps.Bounding, []string{"CAP_CHOWN", "CAP_NET_ADMIN"}) {
		t.Fatalf("Bounding = %v", b.caps.Bounding)
	}
	if !slices.Equal(b.caps.Effective, []string{"CAP_KILL"}) {
		t.Fatalf("Effective = %v", b.caps.Effective)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	b := NewBuilder()
	g := &grant.Grant{
		Subsystem: "cap",
		Args:      []grant.Arg{{Type: grant.ArgName, Value: "net_admin"}},
		Where:     &grant.CompareExpr{},
	}
	if err := b.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	b := NewBuilder()
	g := &grant.Grant{
		Subsystem: "cap",
		Args:      []grant.Arg{{Type: grant.ArgName, Value: "net_admin"}},
		Kwargs:    []grant.Kwarg{{Key: "k", Value: grant.Arg{Type: grant.ArgName, Value: "v"}}},
	}
	if err := b.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsMissingArgs(t *testing.T) {
	if err := NewBuilder().Build(&grant.Grant{Subsystem: "cap"}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsTooManyArgs(t *testing.T) {
	b := NewBuilder()
	g := &grant.Grant{Subsystem: "cap", Args: []grant.Arg{
		{Type: grant.ArgName, Value: "net_admin"},
		{Type: grant.ArgName, Value: "full"},
		{Type: grant.ArgName, Value: "extra"},
	}}
	if err := b.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonNameCapArg(t *testing.T) {
	if err := NewBuilder().Build(&grant.Grant{
		Subsystem: "cap",
		Args:      []grant.Arg{{Type: grant.ArgInt, Value: "1"}},
	}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownCap(t *testing.T) {
	if err := NewBuilder().Build(&grant.Grant{
		Subsystem: "cap",
		Args:      []grant.Arg{{Type: grant.ArgName, Value: "not_a_cap"}},
	}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonNameModeArg(t *testing.T) {
	if err := NewBuilder().Build(&grant.Grant{
		Subsystem: "cap",
		Args: []grant.Arg{
			{Type: grant.ArgName, Value: "net_admin"},
			{Type: grant.ArgInt, Value: "1"},
		},
	}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownMode(t *testing.T) {
	if err := NewBuilder().Build(&grant.Grant{
		Subsystem: "cap",
		Args: []grant.Arg{
			{Type: grant.ArgName, Value: "net_admin"},
			{Type: grant.ArgName, Value: "bogus"},
		},
	}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildEffectiveMode(t *testing.T) {
	b := NewBuilder()
	if err := b.Build(&grant.Grant{
		Subsystem: "cap",
		Args: []grant.Arg{
			{Type: grant.ArgName, Value: "chown"},
			{Type: grant.ArgName, Value: "effective"},
		},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := b.Spec()
	if len(s.Effective) == 0 || len(s.Permitted) == 0 || len(s.Bounding) == 0 {
		t.Fatalf("effective/permitted/bounding should be set: %+v", s)
	}
	if len(s.Inheritable) != 0 || len(s.Ambient) != 0 {
		t.Fatalf("inheritable/ambient should be empty: %+v", s)
	}
}

func TestBuildInheritableMode(t *testing.T) {
	b := NewBuilder()
	if err := b.Build(&grant.Grant{
		Subsystem: "cap",
		Args: []grant.Arg{
			{Type: grant.ArgName, Value: "chown"},
			{Type: grant.ArgName, Value: "inheritable"},
		},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := b.Spec()
	if len(s.Permitted) == 0 || len(s.Inheritable) == 0 || len(s.Ambient) == 0 || len(s.Bounding) == 0 {
		t.Fatalf("permitted/inheritable/ambient/bounding should be set: %+v", s)
	}
	if len(s.Effective) != 0 {
		t.Fatalf("effective should be empty: %+v", s)
	}
}

func TestBuildPermittedMode(t *testing.T) {
	b := NewBuilder()
	if err := b.Build(&grant.Grant{
		Subsystem: "cap",
		Args: []grant.Arg{
			{Type: grant.ArgName, Value: "chown"},
			{Type: grant.ArgName, Value: "permitted"},
		},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := b.Spec()
	if len(s.Permitted) == 0 || len(s.Bounding) == 0 {
		t.Fatalf("permitted/bounding should be set: %+v", s)
	}
	if len(s.Effective) != 0 || len(s.Inheritable) != 0 || len(s.Ambient) != 0 {
		t.Fatalf("effective/inheritable/ambient should be empty: %+v", s)
	}
}

func TestBuildIdempotent(t *testing.T) {
	b := NewBuilder()
	g := &grant.Grant{Subsystem: "cap", Args: []grant.Arg{{Type: grant.ArgName, Value: "net_admin"}}}
	if err := b.Build(g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := b.Build(g); err != nil {
		t.Fatalf("Build re: %v", err)
	}
	if s := b.Spec(); len(s.Effective) != 1 {
		t.Fatalf("Effective len = %d, want 1 after idempotent builds", len(s.Effective))
	}
}

func TestMergeNilIsNoOp(t *testing.T) {
	b := NewBuilder()
	b.apply("net_admin", ModeFull)
	if err := b.Merge(nil); err != nil {
		t.Fatalf("Merge(nil): %v", err)
	}
	if len(b.caps.Effective) != 1 {
		t.Fatal("Merge(nil) mutated the builder")
	}
}
