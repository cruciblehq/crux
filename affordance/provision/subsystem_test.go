package provision

import (
	"errors"
	"testing"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

func nameArg(v string) agl.Arg     { return agl.Arg{Type: agl.ArgName, Value: v} }
func intArg(v string) agl.Arg      { return agl.Arg{Type: agl.ArgInt, Value: v} }
func quantityArg(v string) agl.Arg { return agl.Arg{Type: agl.ArgQuantity, Value: v} }

func newSub() (*Subsystem, *Spec) {
	s := &Spec{}
	return New(s), s
}

func TestNameReturnsProvision(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != subsystem.NameProvision {
		t.Fatalf("Name() = %q, want %q", got, subsystem.NameProvision)
	}
}

func TestSpecValidate(t *testing.T) {
	s := &Spec{CPU: 500, Memory: 1 << 20, Disk: 1 << 30}
	if err := s.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestBuildCPUInt(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("cpu"), intArg("4")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if s.CPU != 4000 {
		t.Fatalf("CPU = %d, want 4000", s.CPU)
	}
}

func TestBuildCPUMillicore(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("cpu"), quantityArg("500m")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if s.CPU != 500 {
		t.Fatalf("CPU = %d, want 500", s.CPU)
	}
}

func TestBuildCPUBadSuffix(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("cpu"), quantityArg("4Gi")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildCPUMillicoreOverflow(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("cpu"), quantityArg("99999999999999999999999m")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildMemorySmallQuantity(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("memory"), quantityArg("5k")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if s.Memory != 5000 {
		t.Fatalf("Memory = %d, want 5000", s.Memory)
	}
}

func TestBuildMemoryQuantity(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("memory"), quantityArg("8Gi")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := uint64(8) * 1024 * 1024 * 1024
	if s.Memory != want {
		t.Fatalf("Memory = %d, want %d", s.Memory, want)
	}
}

func TestBuildDiskInt(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("disk"), intArg("1073741824")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if s.Disk != 1073741824 {
		t.Fatalf("Disk = %d, want 1073741824", s.Disk)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Args:  []agl.Arg{nameArg("cpu"), intArg("4")},
		Where: &agl.CompareExpr{},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Args:   []agl.Arg{nameArg("cpu"), intArg("4")},
		Kwargs: []agl.Kwarg{{Key: "k", Value: nameArg("v")}},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWrongArgCount(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("cpu")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonNameResource(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{intArg("42"), intArg("4")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildUnknownResource(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("gpu"), intArg("4")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildMemoryBadType(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("memory"), nameArg("bigmem")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseByteQuantityUnknownSuffix(t *testing.T) {
	_, err := parseByteQuantity("4zz")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildDiskBadType(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("disk"), nameArg("bigdisk")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildCPUNameArg(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("cpu"), nameArg("lots")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildCPUIntOverflow(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("cpu"), intArg("18446744073709551616")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildMemoryIntOverflow(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("memory"), intArg("18446744073709551616")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParseByteQuantityBadNumber(t *testing.T) {
	_, err := parseByteQuantity("aaaGi")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
