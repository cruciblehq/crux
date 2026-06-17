package mac

import (
	"errors"
	"testing"

	"github.com/cruciblehq/crux/affordance/agl"
)

func newSub() (*Subsystem, *Spec) {
	s := &Spec{}
	return New(s), s
}

func buildSrc(t *testing.T, sub *Subsystem, src string) error {
	t.Helper()
	g, err := agl.Parse(src)
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
	got := &Spec{}
	if len(got.Rules) != 0 {
		t.Fatalf("Rules len = %d, want 0", len(got.Rules))
	}
}

func TestBuildRejectsNoArgs(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "mac"}); !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildRejectsTwoArgs(t *testing.T) {
	sub, _ := newSub()
	if err := sub.Build(&agl.Model{
		Subsystem: "mac",
		Args: []agl.Arg{
			{Type: agl.ArgName, Value: "file_open"},
			{Type: agl.ArgName, Value: "extra"},
		},
	}); err == nil {
		t.Fatal("expected error for two args")
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

func TestNameReturnsMAC(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != "mac" {
		t.Fatalf("Name() = %q, want %q", got, "mac")
	}
}

func TestBuildWithNotWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, `.mac file_open where not file.path = "/x"`); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "not" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildWithBitTestWhere(t *testing.T) {
	sub, s := newSub()
	if err := buildSrc(t, sub, `.mac file_open where task.uid & 0x4`); err != nil {
		t.Fatal(err)
	}
	r := s.Rules[0]
	if r.Where == nil || r.Where.Type != "bittest" {
		t.Fatalf("where = %#v", r.Where)
	}
}

func TestBuildLikeOnNumericField(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where task.uid like "/x"`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBetweenOnStringField(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where file.path between 1 and 2`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildRejectsNonNameHookArg(t *testing.T) {
	sub, _ := newSub()
	g := &agl.Model{Args: []agl.Arg{{Type: agl.ArgInt, Value: "1"}}}
	if err := sub.Build(g); !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBinaryLeftFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where bogus.field = 1 and task.uid = 0`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBinaryRightFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where task.uid = 0 and bogus.field = 1`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildUnaryFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where not bogus.field = 1`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildCompareRHSFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where task.uid = bogus.field`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildLikeUnknownField(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where bogus.field like "/x"`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBetweenFieldFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where bogus.field between 1 and 100`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBetweenLowFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where task.uid between bogus.field and 100`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBetweenHighFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where task.uid between 0 and bogus.field`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBitTestUnknownField(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where bogus.field & 7`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildBitTestStringField(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where file.path & 7`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildInFieldFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where bogus.field in (0)`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildInValueFails(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where task.uid in (bogus.field)`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestBuildInValueTypeMismatch(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, `.mac file_open where task.uid in (file.path)`)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

// customExpr satisfies agl.Expr but is not one of the seven known node types.
type customExpr struct{}

func (e *customExpr) String() string { return "custom" }

func TestTranslateExprUnknownType(t *testing.T) {
	hook := catalog().LookupHook("file_open")
	if hook == nil {
		t.Fatal("file_open hook not found")
	}
	_, err := translateExpr(&customExpr{}, hook)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestTranslateOperandValueVar(t *testing.T) {
	hook := catalog().LookupHook("file_open")
	_, err := translateOperand(agl.Operand{Value: agl.Value{Type: agl.ValueVar, Str: "x"}}, hook)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestTranslateOperandUnknownType(t *testing.T) {
	hook := catalog().LookupHook("file_open")
	_, err := translateOperand(agl.Operand{Value: agl.Value{Type: 99}}, hook)
	if !errors.Is(err, ErrCompile) {
		t.Fatalf("err = %v, want ErrCompile", err)
	}
}

func TestRequireFieldTypeNonField(t *testing.T) {
	hook := catalog().LookupHook("file_open")
	op := agl.Operand{IsField: false, Value: agl.Value{Type: agl.ValueInt, Int: 0}}
	if err := requireFieldType(op, hook, TypeString, "'like'"); err != nil {
		t.Fatalf("expected nil for non-field operand, got %v", err)
	}
}

func TestResolveTypeUnknownField(t *testing.T) {
	hook := catalog().LookupHook("file_open")
	op := agl.Operand{IsField: true, Field: "bogus.field"}
	if got := resolveType(op, hook); got != nil {
		t.Fatalf("expected nil for unknown field, got %v", *got)
	}
}

func TestResolveTypeVarReturnsNil(t *testing.T) {
	hook := catalog().LookupHook("file_open")
	op := agl.Operand{Value: agl.Value{Type: agl.ValueVar, Str: "x"}}
	if got := resolveType(op, hook); got != nil {
		t.Fatalf("expected nil for var operand, got %v", *got)
	}
}

func TestCheckTypeCompatNilType(t *testing.T) {
	hook := catalog().LookupHook("file_open")
	// Left has a resolvable type; right has a variable reference (nil type).
	left := agl.Operand{IsField: true, Field: "task.uid"}
	right := agl.Operand{Value: agl.Value{Type: agl.ValueVar, Str: "x"}}
	if err := checkTypeCompat(left, right, hook); err != nil {
		t.Fatalf("expected nil for unresolvable right operand type, got %v", err)
	}
}
