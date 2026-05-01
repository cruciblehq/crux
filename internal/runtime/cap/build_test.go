package cap

import (
	"errors"
	"slices"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

// Returns a Subsystem with an empty caps section.
func newSub() (*Subsystem, *specs.LinuxCapabilities) {
	caps := &specs.LinuxCapabilities{}
	return New(caps), caps
}

// Wraps a caps section into a unified spec for Merge inputs.
func wrap(c *specs.LinuxCapabilities) shared.Spec {
	return shared.Spec{OCI: &specs.Spec{Process: &specs.Process{Capabilities: c}}}
}

func TestApplyFullMode(t *testing.T) {
	caps := &specs.LinuxCapabilities{}
	if err := apply(caps, "net_admin", modeFull); err != nil {
		t.Fatalf("apply: %v", err)
	}
	want := []string{"CAP_NET_ADMIN"}
	for name, got := range map[string][]string{
		"effective":   caps.Effective,
		"permitted":   caps.Permitted,
		"inheritable": caps.Inheritable,
		"bounding":    caps.Bounding,
		"ambient":     caps.Ambient,
	} {
		if !slices.Equal(got, want) {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
}

func TestMergeUnionsSets(t *testing.T) {
	sub, dst := newSub()
	if err := apply(dst, "chown", modeBound); err != nil {
		t.Fatalf("apply: %v", err)
	}
	src := &specs.LinuxCapabilities{
		Bounding:  []string{"CAP_CHOWN", "CAP_NET_ADMIN"},
		Effective: []string{"CAP_KILL"},
	}
	if err := sub.Merge(wrap(src)); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !slices.Equal(dst.Bounding, []string{"CAP_CHOWN", "CAP_NET_ADMIN"}) {
		t.Fatalf("Bounding = %v", dst.Bounding)
	}
	if !slices.Equal(dst.Effective, []string{"CAP_KILL"}) {
		t.Fatalf("Effective = %v", dst.Effective)
	}
}

func TestMergeNilCapIsNoOp(t *testing.T) {
	sub, dst := newSub()
	if err := apply(dst, "net_admin", modeFull); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if err := sub.Merge(shared.Spec{}); err != nil {
		t.Fatalf("Merge(empty): %v", err)
	}
	if len(dst.Effective) != 1 {
		t.Fatal("Merge(empty) mutated the spec")
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{
		Subsystem: "cap",
		Args:      []grant.Arg{{Type: grant.ArgName, Value: "net_admin"}},
		Where:     &grant.CompareExpr{},
	}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{
		Subsystem: "cap",
		Args:      []grant.Arg{{Type: grant.ArgName, Value: "net_admin"}},
		Kwargs:    []grant.Kwarg{{Key: "k", Value: grant.Arg{Type: grant.ArgName, Value: "v"}}},
	}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsMissingArgs(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(grant.Grant{Subsystem: "cap"}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsTooManyArgs(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "cap", Args: []grant.Arg{
		{Type: grant.ArgName, Value: "net_admin"},
		{Type: grant.ArgName, Value: "full"},
		{Type: grant.ArgName, Value: "extra"},
	}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownCap(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(grant.Grant{
		Subsystem: "cap",
		Args:      []grant.Arg{{Type: grant.ArgName, Value: "not_a_cap"}},
	}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownMode(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(grant.Grant{
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
	sub, caps := newSub()
	if err := sub.Build(grant.Grant{
		Subsystem: "cap",
		Args: []grant.Arg{
			{Type: grant.ArgName, Value: "chown"},
			{Type: grant.ArgName, Value: "effective"},
		},
	}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(caps.Effective) == 0 || len(caps.Permitted) == 0 || len(caps.Bounding) == 0 {
		t.Fatalf("effective/permitted/bounding should be set: %+v", caps)
	}
	if len(caps.Inheritable) != 0 || len(caps.Ambient) != 0 {
		t.Fatalf("inheritable/ambient should be empty: %+v", caps)
	}
}

func TestBuildIdempotent(t *testing.T) {
	sub, caps := newSub()
	g := grant.Grant{Subsystem: "cap", Args: []grant.Arg{{Type: grant.ArgName, Value: "net_admin"}}}
	if err := sub.Build(g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := sub.Build(g); err != nil {
		t.Fatalf("Build re: %v", err)
	}
	if len(caps.Effective) != 1 {
		t.Fatalf("Effective len = %d, want 1 after idempotent builds", len(caps.Effective))
	}
}

func TestNameReturnsCap(t *testing.T) {
	sub, _ := newSub()
	if sub.Name() != shared.NameCap {
		t.Fatalf("Name() = %q, want %q", sub.Name(), shared.NameCap)
	}
}

func TestApplyInheritableMode(t *testing.T) {
	caps := &specs.LinuxCapabilities{}
	if err := apply(caps, "net_admin", modeInheritable); err != nil {
		t.Fatalf("apply: %v", err)
	}
	want := []string{"CAP_NET_ADMIN"}
	for name, got := range map[string][]string{
		"permitted":   caps.Permitted,
		"inheritable": caps.Inheritable,
		"ambient":     caps.Ambient,
		"bounding":    caps.Bounding,
	} {
		if !slices.Equal(got, want) {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
	if len(caps.Effective) != 0 {
		t.Errorf("effective = %v, want empty", caps.Effective)
	}
}

func TestApplyPermittedMode(t *testing.T) {
	caps := &specs.LinuxCapabilities{}
	if err := apply(caps, "net_admin", modePermitted); err != nil {
		t.Fatalf("apply: %v", err)
	}
	want := []string{"CAP_NET_ADMIN"}
	if !slices.Equal(caps.Permitted, want) {
		t.Errorf("permitted = %v, want %v", caps.Permitted, want)
	}
	if !slices.Equal(caps.Bounding, want) {
		t.Errorf("bounding = %v, want %v", caps.Bounding, want)
	}
	if len(caps.Effective) != 0 || len(caps.Inheritable) != 0 || len(caps.Ambient) != 0 {
		t.Errorf("effective/inheritable/ambient should be empty: %+v", caps)
	}
}

func TestApplyBoundMode(t *testing.T) {
	caps := &specs.LinuxCapabilities{}
	if err := apply(caps, "net_admin", modeBound); err != nil {
		t.Fatalf("apply: %v", err)
	}
	want := []string{"CAP_NET_ADMIN"}
	if !slices.Equal(caps.Bounding, want) {
		t.Errorf("bounding = %v, want %v", caps.Bounding, want)
	}
	if len(caps.Effective) != 0 || len(caps.Permitted) != 0 || len(caps.Inheritable) != 0 || len(caps.Ambient) != 0 {
		t.Errorf("only bounding should be set: %+v", caps)
	}
}
