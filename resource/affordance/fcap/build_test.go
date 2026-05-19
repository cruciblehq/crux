package fcap

import (
	"errors"
	"slices"
	"testing"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/resource/affordance/agl"
)

// Helper to create a name-typed argument for testing.
func nameArg(v string) agl.Arg {
	return agl.Arg{Type: agl.ArgName, Value: v}
}

// Helper to create a string-typed argument for testing.
func strArg(v string) agl.Arg {
	return agl.Arg{Type: agl.ArgStrASCII, Value: v}
}

// Helper to create a Subsystem with a new spec for testing.
func newSub() (*Subsystem, *manifest.FcapSpec) {
	s := &manifest.FcapSpec{}
	return New(s), s
}

func TestBuildEffectiveSetsPermittedAndEffective(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("net_admin"), nameArg("effective"), strArg("/usr/bin/foo")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	e, ok := s.Entries["/usr/bin/foo"]
	if !ok {
		t.Fatal("entry missing")
	}
	if !e.Effective {
		t.Fatal("effective bit not set")
	}
	if !slices.Equal(e.Permitted, []string{"net_admin"}) {
		t.Fatalf("Permitted = %v", e.Permitted)
	}
}

func TestBuildInheritableOnly(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("chown"), nameArg("inheritable"), strArg("/bin/sh")}}
	if err := sub.Build(&g); err != nil {
		t.Fatal(err)
	}
	e := s.Entries["/bin/sh"]
	if e.Effective {
		t.Fatal("effective bit should not be set")
	}
	if !slices.Equal(e.Inheritable, []string{"chown"}) {
		t.Fatalf("Inheritable = %v", e.Inheritable)
	}
}

func TestBuildSameGrantConflict(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("net_admin"), nameArg("effective"), strArg("/usr/bin/x")}}
	if err := sub.Build(&g); err != nil {
		t.Fatal(err)
	}
	err := sub.Build(&g)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("net_admin"), nameArg("effective"), strArg("/x")}, Where: &agl.CompareExpr{}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWrongArgCount(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("net_admin")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsUnknownCap(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("bogus"), nameArg("effective"), strArg("/x")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsUnknownMode(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("chown"), nameArg("bogus"), strArg("/x")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsRelativePath(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("chown"), nameArg("effective"), strArg("foo/bar")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsTrailingSlash(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("chown"), nameArg("effective"), strArg("/usr/bin/")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsUnclean(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("chown"), nameArg("effective"), strArg("/usr//bin/x")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsNUL(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Subsystem: "fcap", Args: []agl.Arg{nameArg("chown"), nameArg("effective"), strArg("/x\x00y")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

