package kernel

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

func TestBuildSingleFeature(t *testing.T) {
	sub, vm := newSub()
	if err := sub.Build(parse(t, ".kernel config FUSE_FS")); err != nil {
		t.Fatal(err)
	}
	if len(vm.Features) != 1 || vm.Features[0] != "FUSE_FS" {
		t.Errorf("Features = %v, want [FUSE_FS]", vm.Features)
	}
}

func TestBuildMultipleFeatures(t *testing.T) {
	sub, vm := newSub()
	if err := sub.Build(parse(t, ".kernel config FUSE_FS")); err != nil {
		t.Fatal(err)
	}
	if err := sub.Build(parse(t, ".kernel config NETFILTER")); err != nil {
		t.Fatal(err)
	}
	if len(vm.Features) != 2 {
		t.Fatalf("Features = %v, want [FUSE_FS NETFILTER]", vm.Features)
	}
}

func TestBuildModule(t *testing.T) {
	sub, vm := newSub()
	if err := sub.Build(parse(t, ".kernel module fuse")); err != nil {
		t.Fatal(err)
	}
	if len(vm.Modules) != 1 || vm.Modules[0] != "fuse" {
		t.Errorf("Modules = %v, want [fuse]", vm.Modules)
	}
}

func TestBuildVersion(t *testing.T) {
	sub, vm := newSub()
	if err := sub.Build(parse(t, `.kernel version "5.15"`)); err != nil {
		t.Fatal(err)
	}
	if len(vm.Versions) != 1 || vm.Versions[0] != "5.15" {
		t.Errorf("Versions = %v, want [5.15]", vm.Versions)
	}
}

func TestBuildLSM(t *testing.T) {
	sub, vm := newSub()
	if err := sub.Build(parse(t, ".kernel lsm apparmor")); err != nil {
		t.Fatal(err)
	}
	if len(vm.LSMs) != 1 || vm.LSMs[0] != "apparmor" {
		t.Errorf("LSMs = %v, want [apparmor]", vm.LSMs)
	}
}

func TestBuildHW(t *testing.T) {
	sub, vm := newSub()
	if err := sub.Build(parse(t, ".kernel hw sgx")); err != nil {
		t.Fatal(err)
	}
	if len(vm.HWFeatures) != 1 || vm.HWFeatures[0] != "sgx" {
		t.Errorf("HWFeatures = %v, want [sgx]", vm.HWFeatures)
	}
}

func TestBuildRejectsUnknownType(t *testing.T) {
	sub, _ := newSub()
	g := &agl.Model{
		Subsystem: "kernel",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "unknown"}, {Type: agl.ArgName, Value: "x"}},
	}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Errorf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWhereClause(t *testing.T) {
	sub, _ := newSub()
	g := parse(t, ".kernel config FUSE_FS")
	g.Where = &agl.CompareExpr{}
	err := sub.Build(g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Errorf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	g := parse(t, ".kernel config FUSE_FS")
	g.Kwargs = []agl.Kwarg{{Key: "mode", Value: agl.Arg{Value: "y"}}}
	err := sub.Build(g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Errorf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsMissingArg(t *testing.T) {
	sub, _ := newSub()
	g := &agl.Model{Subsystem: "kernel"}
	err := sub.Build(g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Errorf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestKey(t *testing.T) {
	sub, _ := newSub()
	g := parse(t, ".kernel config FUSE_FS")
	if got := sub.Key(g); got != "config:FUSE_FS" {
		t.Errorf("Key = %q, want %q", got, "config:FUSE_FS")
	}
}

func TestName(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != subsystem.NameKernel {
		t.Fatalf("Name() = %q, want %q", got, subsystem.NameKernel)
	}
}

func TestBuildBoot(t *testing.T) {
	sub, vm := newSub()
	if err := sub.Build(parse(t, ".kernel boot quiet")); err != nil {
		t.Fatal(err)
	}
	if len(vm.BootParams) != 1 || vm.BootParams[0] != "quiet" {
		t.Errorf("BootParams = %v, want [quiet]", vm.BootParams)
	}
}

func TestKeyEmptyWhenNoValue(t *testing.T) {
	sub, _ := newSub()
	g := &agl.Model{Args: []agl.Arg{{Type: agl.ArgName, Value: "config"}}}
	if got := sub.Key(g); got != "" {
		t.Fatalf("Key() = %q, want empty string", got)
	}
}

func TestBuildRejectsNonNameType(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".kernel 5 x"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonStringValue(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".kernel config 5"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
