package cap

import (
	"errors"
	"slices"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

// Returns a Subsystem with an empty caps section.
func newSub() (*Subsystem, *specs.LinuxCapabilities) {
	caps := &specs.LinuxCapabilities{}
	return New(caps), caps
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

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Subsystem: "cap",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "net_admin"}},
		Where:     &agl.CompareExpr{},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Subsystem: "cap",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "net_admin"}},
		Kwargs:    []agl.Kwarg{{Key: "k", Value: agl.Arg{Type: agl.ArgName, Value: "v"}}},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsMissingArgs(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "cap"}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsTooManyArgs(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "cap", Args: []agl.Arg{
		{Type: agl.ArgName, Value: "net_admin"},
		{Type: agl.ArgName, Value: "full"},
		{Type: agl.ArgName, Value: "extra"},
	}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownCap(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&agl.Model{
		Subsystem: "cap",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "not_a_cap"}},
	}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownMode(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&agl.Model{
		Subsystem: "cap",
		Args: []agl.Arg{
			{Type: agl.ArgName, Value: "net_admin"},
			{Type: agl.ArgName, Value: "bogus"},
		},
	}); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildEffectiveMode(t *testing.T) {
	sub, caps := newSub()
	if err := sub.Build(&agl.Model{
		Subsystem: "cap",
		Args: []agl.Arg{
			{Type: agl.ArgName, Value: "chown"},
			{Type: agl.ArgName, Value: "effective"},
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

func TestNameReturnsCap(t *testing.T) {
	sub, _ := newSub()
	if sub.Name() != subsystem.NameCap {
		t.Fatalf("Name() = %q, want %q", sub.Name(), subsystem.NameCap)
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

func TestKeyReturnsFirstArg(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{{Type: agl.ArgName, Value: "net_admin"}}}
	if got := sub.Key(&g); got != "net_admin" {
		t.Fatalf("Key() = %q, want %q", got, "net_admin")
	}
}

func TestKeyEmptyWhenNoArgs(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Key(&agl.Model{}); got != "" {
		t.Fatalf("Key() = %q, want empty", got)
	}
}

func TestParseRejectsNonNameCapArg(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Subsystem: "cap",
		Args:      []agl.Arg{{Type: agl.ArgInt, Value: "42"}},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseRejectsNonNameModeArg(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Subsystem: "cap",
		Args: []agl.Arg{
			{Type: agl.ArgName, Value: "net_admin"},
			{Type: agl.ArgInt, Value: "42"},
		},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestAddCapIdempotent(t *testing.T) {
	dst := []string{}
	if !addCap(&dst, "CAP_NET_ADMIN") {
		t.Fatal("first addCap should return true")
	}
	if addCap(&dst, "CAP_NET_ADMIN") {
		t.Fatal("second addCap should return false (already present)")
	}
	if len(dst) != 1 {
		t.Fatalf("len(dst) = %d, want 1", len(dst))
	}
}
