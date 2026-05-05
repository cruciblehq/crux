package mac

import (
	"errors"
	"testing"

	"github.com/cruciblehq/crux/internal/aegis"
	"github.com/cruciblehq/crux/internal/runtime/shared"
	"github.com/cruciblehq/crux/internal/runtime/shared/macspec"
)

func newSub() (*Subsystem, *macspec.Spec) {
	s := macspec.NewSpec()
	return New(s), s
}

func wrap(s *macspec.Spec) shared.Spec { return shared.Spec{MAC: s} }

func buildSrc(t *testing.T, sub *Subsystem, src string) error {
	t.Helper()
	g, err := aegis.Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	return sub.Build(g)
}

func TestBuildSimpleHook(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, ".mac file_open"); err != nil {
		t.Fatal(err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("rules = %#v", s)
	}
	r := s.Rules[0]
	if r.Hook != "file_open" || r.Where != nil {
		t.Fatalf("rule = %#v", r)
	}
}

func TestBuildWithWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, ".mac file_open where file.path = \"/etc/hosts\""); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "cmp" || r.Where.Op != "=" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildUnknownHook(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, ".mac not_a_hook")
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildUnknownField(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, ".mac file_open where bogus.field = 1")
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildTypeMismatch(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, ".mac file_open where file.path = 1")
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildDeduplicates(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, ".mac file_open"); err != nil {
		t.Fatal(err)
	}
	if err := buildSrc(t, sub, ".mac file_open"); err != nil {
		t.Fatal(err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("rules = %#v", s.Rules)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".mac file_open mode=allow"); err == nil {
		t.Fatal("expected error for kwargs")
	}
}

func TestSpecEmptyWhenNoGrants(t *testing.T) {
	got := macspec.NewSpec()
	if got == nil {
		t.Fatal("NewSpec() = nil")
	}
	if len(got.Rules) != 0 {
		t.Fatalf("Rules len = %d, want 0", len(got.Rules))
	}
}

func TestBuildRejectsNoArgs(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&aegis.Model{Subsystem: "mac"}); !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildRejectsTwoArgs(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&aegis.Model{
		Subsystem: "mac",
		Args: []aegis.Arg{
			{Type: aegis.ArgName, Value: "file_open"},
			{Type: aegis.ArgName, Value: "extra"},
		},
	}); err == nil {
		t.Fatal("expected error for two args")
	}
}

func TestMergeNilIsNoOp(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, ".mac file_open"); err != nil {
		t.Fatal(err)
	}
	if err := sub.Merge(shared.Spec{}); err != nil {
		t.Fatalf("Merge(empty): %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatal("Merge(empty) mutated state")
	}
}

func TestMergeAddsRules(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, ".mac file_open"); err != nil {
		t.Fatal(err)
	}
	otherSub, other := newSub()
	if err := buildSrc(t, otherSub, ".mac file_permission"); err != nil {
		t.Fatal(err)
	}
	if err := sub.Merge(wrap(other)); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(s.Rules) != 2 {
		t.Fatalf("Rules len = %d, want 2", len(s.Rules))
	}
}

func TestBuildWithInWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, `.mac file_open where file.path in ("/etc/passwd", "/etc/shadow")`); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "in" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildWithLikeWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, `.mac file_open where file.path like "/etc/**"`); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "like" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildWithAndWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, `.mac file_open where file.path = "/etc/hosts" and task.uid = 0`); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "and" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildWithOrWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, `.mac file_open where file.path = "/etc/hosts" or task.uid = 0`); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "or" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildWithBetweenWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, `.mac file_open where task.uid between 1000 and 65534`); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "between" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildSleepableFieldOnNonSleepableHookFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_permission where file.ima_hash = "sha256:deadbeef"`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}
