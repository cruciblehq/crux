package fcap

import (
	"errors"
	"slices"
	"testing"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
	fcapspec "github.com/cruciblehq/crux/internal/runtime/shared/fcap"
)

// Helper to create a name-typed argument for testing.
func nameArg(v string) grant.Arg {
	return grant.Arg{Type: grant.ArgName, Value: v}
}

// Helper to create a string-typed argument for testing.
func strArg(v string) grant.Arg {
	return grant.Arg{Type: grant.ArgStrASCII, Value: v}
}

// Helper to create a Subsystem with a new spec for testing.
func newSub() (*Subsystem, *fcapspec.Spec) {
	s := fcapspec.NewSpec()
	return New(s), s
}

// Helper to wrap an fcap spec in a shared spec for Merge testing.
func wrap(s *fcapspec.Spec) shared.Spec {
	return shared.Spec{Fcap: s}
}

func TestBuildEffectiveSetsPermittedAndEffective(t *testing.T) {
	sub, s := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("net_admin"), nameArg("effective"), strArg("/usr/bin/foo")}}
	if err := sub.Build(g); err != nil {
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
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("chown"), nameArg("inheritable"), strArg("/bin/sh")}}
	if err := sub.Build(g); err != nil {
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

func TestBuildIdempotent(t *testing.T) {
	sub, s := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("net_admin"), nameArg("effective"), strArg("/usr/bin/x")}}
	if err := sub.Build(g); err != nil {
		t.Fatal(err)
	}
	if err := sub.Build(g); err != nil {
		t.Fatal(err)
	}
	if got := len(s.Entries["/usr/bin/x"].Permitted); got != 1 {
		t.Fatalf("Permitted len = %d, want 1", got)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("net_admin"), nameArg("effective"), strArg("/x")}, Where: &grant.CompareExpr{}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWrongArgCount(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("net_admin")}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsUnknownCap(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("bogus"), nameArg("effective"), strArg("/x")}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsUnknownMode(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("chown"), nameArg("bogus"), strArg("/x")}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsRelativePath(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("chown"), nameArg("effective"), strArg("foo/bar")}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsTrailingSlash(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("chown"), nameArg("effective"), strArg("/usr/bin/")}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsUnclean(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("chown"), nameArg("effective"), strArg("/usr//bin/x")}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsNUL(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("chown"), nameArg("effective"), strArg("/x\x00y")}}
	if err := sub.Build(g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestMergeUnions(t *testing.T) {
	dstSub, dst := newSub()
	srcSub, src := newSub()
	if err := srcSub.Build(grant.Grant{Subsystem: "fcap", Args: []grant.Arg{nameArg("net_admin"), nameArg("effective"), strArg("/x")}}); err != nil {
		t.Fatal(err)
	}
	if err := dstSub.Merge(wrap(src)); err != nil {
		t.Fatal(err)
	}
	if _, ok := dst.Entries["/x"]; !ok {
		t.Fatal("merge did not import entry")
	}
}

func TestMergeNilIsNoOp(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Merge(shared.Spec{}); err != nil {
		t.Fatalf("Merge(empty): %v", err)
	}
}
