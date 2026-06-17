package mount

import (
	"errors"
	"testing"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func parse(t *testing.T, src string) *agl.Model {
	t.Helper()
	g, err := agl.Parse(src)
	if err != nil {
		t.Fatalf("agl.Parse(%q): %v", src, err)
	}
	return g
}

func newSub() (*Subsystem, *[]specs.Mount) {
	s := make([]specs.Mount, 0)
	return New(&s), &s
}

func TestBuildTmpfsDefaults(t *testing.T) {
	sub, mounts := newSub()
	if err := sub.Build(parse(t, ".mount tmpfs /tmp")); err != nil {
		t.Fatal(err)
	}
	if len(*mounts) != 1 {
		t.Fatalf("want 1 mount, got %d", len(*mounts))
	}
	m := (*mounts)[0]
	if m.Type != "tmpfs" || m.Source != "tmpfs" || m.Destination != "/tmp" {
		t.Errorf("unexpected mount: %+v", m)
	}
	hasMode := false
	for _, o := range m.Options {
		if o == "mode=1777" {
			hasMode = true
		}
	}
	if !hasMode {
		t.Errorf("expected mode=1777 in options: %v", m.Options)
	}
}

func TestBuildTmpfsWithSize(t *testing.T) {
	sub, mounts := newSub()
	if err := sub.Build(parse(t, ".mount tmpfs /run/cache size=64Mi")); err != nil {
		t.Fatal(err)
	}
	m := (*mounts)[0]
	hasSize := false
	for _, o := range m.Options {
		if o == "size=67108864" {
			hasSize = true
		}
	}
	if !hasSize {
		t.Errorf("expected size=67108864 in options: %v", m.Options)
	}
}

func TestBuildTmpfsWithMode(t *testing.T) {
	sub, mounts := newSub()
	if err := sub.Build(parse(t, ".mount tmpfs /tmp mode=0755")); err != nil {
		t.Fatal(err)
	}
	m := (*mounts)[0]
	hasMode := false
	for _, o := range m.Options {
		if o == "mode=0755" {
			hasMode = true
		}
	}
	if !hasMode {
		t.Errorf("expected mode=0755 in options: %v", m.Options)
	}
}

func TestBuildProc(t *testing.T) {
	sub, mounts := newSub()
	if err := sub.Build(parse(t, ".mount proc /proc")); err != nil {
		t.Fatal(err)
	}
	m := (*mounts)[0]
	if m.Type != "proc" || m.Destination != "/proc" {
		t.Errorf("unexpected mount: %+v", m)
	}
}

func TestBuildRejectsUnknownType(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".mount bind /data"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildTmpfsSizeInteger(t *testing.T) {
	sub, mounts := newSub()
	if err := sub.Build(parse(t, ".mount tmpfs /run/cache size=4096")); err != nil {
		t.Fatal(err)
	}
	m := (*mounts)[0]
	hasSize := false
	for _, o := range m.Options {
		if o == "size=4096" {
			hasSize = true
		}
	}
	if !hasSize {
		t.Errorf("expected size=4096 in options: %v", m.Options)
	}
}

func TestBuildTmpfsSizeSingleCharSuffix(t *testing.T) {
	sub, mounts := newSub()
	if err := sub.Build(parse(t, ".mount tmpfs /run/cache size=1k")); err != nil {
		t.Fatal(err)
	}
	m := (*mounts)[0]
	hasSize := false
	for _, o := range m.Options {
		if o == "size=1000" {
			hasSize = true
		}
	}
	if !hasSize {
		t.Errorf("expected size=1000 in options: %v", m.Options)
	}
}

func TestBuildRejectsSubUnitSize(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".mount tmpfs /tmp size=64m"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonOctalMode(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".mount tmpfs /tmp mode=999"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonIntMode(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "mount",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "tmpfs"}, {Type: agl.ArgStrASCII, Value: "/tmp"}},
		Kwargs:    []agl.Kwarg{{Key: "mode", Value: agl.Arg{Type: agl.ArgName, Value: "rwx"}}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonQuantitySize(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "mount",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "tmpfs"}, {Type: agl.ArgStrASCII, Value: "/tmp"}},
		Kwargs:    []agl.Kwarg{{Key: "size", Value: agl.Arg{Type: agl.ArgName, Value: "big"}}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsBadIntegerSize(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "mount",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "tmpfs"}, {Type: agl.ArgStrASCII, Value: "/tmp"}},
		Kwargs:    []agl.Kwarg{{Key: "size", Value: agl.Arg{Type: agl.ArgInt, Value: "notanumber"}}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsOverflowingQuantitySize(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "mount",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "tmpfs"}, {Type: agl.ArgStrASCII, Value: "/tmp"}},
		Kwargs:    []agl.Kwarg{{Key: "size", Value: agl.Arg{Type: agl.ArgQuantity, Value: "99999999999999999999Gi"}}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsRelativeDestination(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "mount",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "tmpfs"}, {Type: agl.ArgStrASCII, Value: "tmp/cache"}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsSizeOnNonTmpfs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{
		Subsystem: "mount",
		Args:      []agl.Arg{{Type: agl.ArgName, Value: "proc"}, {Type: agl.ArgStrASCII, Value: "/proc"}},
		Kwargs:    []agl.Kwarg{{Key: "size", Value: agl.Arg{Type: agl.ArgName, Value: "64m"}}},
	})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWhereClause(t *testing.T) {
	sub, _ := newSub()
	g := parse(t, ".mount tmpfs /tmp")
	g.Where = &agl.CompareExpr{}
	err := sub.Build(g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestName(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != subsystem.NameMount {
		t.Fatalf("Name() = %q, want %q", got, subsystem.NameMount)
	}
}

func TestBuildRejectsWrongArgCount(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".mount tmpfs"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonNameType(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".mount 5 /tmp"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonPathDest(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".mount tmpfs 5"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownTmpfsKwarg(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parse(t, ".mount tmpfs /tmp foo=bar"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
