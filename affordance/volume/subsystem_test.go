package volume

import (
	"errors"
	"testing"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

func parse(t *testing.T, src string) *agl.Model {
	t.Helper()
	g, err := agl.Parse(src)
	if err != nil {
		t.Fatalf("agl.Parse(%q): %v", src, err)
	}
	return g
}

func newSub() (*Subsystem, *Spec) {
	spec := &Spec{}
	return New(spec), spec
}

func TestBuildDefaultsReadOnly(t *testing.T) {
	sub, spec := newSub()
	if err := sub.Build(parse(t, ".volume /data")); err != nil {
		t.Fatal(err)
	}
	if len(spec.Mounts) != 1 {
		t.Fatalf("want 1 mount, got %d", len(spec.Mounts))
	}
	m := spec.Mounts[0]
	if m.Destination != "/data" || !m.ReadOnly {
		t.Errorf("unexpected mount: %+v", m)
	}
}

func TestBuildReadOnly(t *testing.T) {
	sub, spec := newSub()
	if err := sub.Build(parse(t, ".volume /data r")); err != nil {
		t.Fatal(err)
	}
	m := spec.Mounts[0]
	if !m.ReadOnly {
		t.Errorf("expected ReadOnly=true, got %+v", m)
	}
}

func TestBuildExplicitReadWrite(t *testing.T) {
	sub, spec := newSub()
	if err := sub.Build(parse(t, ".volume /data rw")); err != nil {
		t.Fatal(err)
	}
	m := spec.Mounts[0]
	if m.ReadOnly {
		t.Errorf("expected ReadOnly=false, got %+v", m)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "volume",
		Args:      []agl.Arg{{Type: agl.ArgStrASCII, Value: "/data"}},
		Kwargs:    []agl.Kwarg{{Key: "k", Value: agl.Arg{Type: agl.ArgName, Value: "v"}}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWhereClause(t *testing.T) {
	sub, _ := newSub()
	g := parse(t, ".volume /data")
	g.Where = &agl.CompareExpr{}
	err := sub.Build(g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsBadSecondArg(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "volume",
		Args: []agl.Arg{
			{Type: agl.ArgStrASCII, Value: "/data"},
			{Type: agl.ArgName, Value: "rwx"},
		},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsRelativeDestination(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "volume",
		Args:      []agl.Arg{{Type: agl.ArgStrASCII, Value: "data/cache"}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonCleanDestination(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "volume",
		Args:      []agl.Arg{{Type: agl.ArgStrASCII, Value: "/data/../etc"}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestMountValidate(t *testing.T) {
	if err := (&Mount{Destination: "/data"}).Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestMountValidateRejectsEmptyDestination(t *testing.T) {
	err := (&Mount{}).Validate()
	if !errors.Is(err, ErrInvalidMount) {
		t.Fatalf("err = %v, want ErrInvalidMount", err)
	}
}

func TestName(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != subsystem.NameVolume {
		t.Fatalf("Name() = %q, want %q", got, subsystem.NameVolume)
	}
}

func TestBuildRejectsWrongArgCount(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".volume /data r extra"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonPathFirstArg(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "volume",
		Args:      []agl.Arg{{Type: agl.ArgInt, Value: "5"}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestSpecValidate(t *testing.T) {
	s := &Spec{Mounts: []Mount{{Destination: "/data"}}}
	if err := s.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestSpecValidateRejectsInvalidMount(t *testing.T) {
	s := &Spec{Mounts: []Mount{{Destination: ""}}}
	if err := s.Validate(); !errors.Is(err, ErrInvalidMount) {
		t.Fatalf("err = %v, want ErrInvalidMount", err)
	}
}
