package mac

import (
	"github.com/cruciblehq/crux/internal/crex"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
	macspec "github.com/cruciblehq/crux/internal/runtime/shared/mac"
)

// Implementation for MAC (LSM) grants.
//
// Holds a pointer to the mac section of the unified spec.
type Subsystem struct {
	spec *macspec.Spec // Pointer to the unified spec's mac section.
}

// Returns a Subsystem wired to mutate spec.
func New(spec *macspec.Spec) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the mac subsystem identifier.
func (s *Subsystem) Name() shared.Name {
	return shared.NameMAC
}

// Applies a parsed grant to the wired-in section.
//
// The grant has the form ".mac HOOK [where EXPR]".
func (s *Subsystem) Build(g grant.Grant) error {
	if err := check(&g); err != nil {
		return err
	}
	allow, err := parse(&g)
	if err != nil {
		return err
	}
	s.spec.Apply(allow)
	return nil
}

// Folds the mac section of src into the wired-in section.
//
// A nil src.MAC is a no-op.
func (s *Subsystem) Merge(src shared.Spec) error {
	if src.MAC == nil {
		return nil
	}
	s.spec.MergeSpec(src.MAC)
	return nil
}

// Validates the grant's structural shape against what the mac subsystem accepts.
func check(g *grant.Grant) error {
	if len(g.Kwargs) != 0 {
		return crex.Wrapf(ErrCompile, "unexpected keyword arguments in mac expression")
	}
	if len(g.Args) != 1 {
		return crex.Wrapf(ErrCompile, "wrong number of arguments in mac expression")
	}
	return nil
}

// Extracts and validates the grant's content into an Allow rule.
//
// Validates that the hook exists and that the where clause, if present, is
// well-formed and references only fields available on the hook. Does not check
// for semantic validity of the where clause beyond field reference checks;
// the resulting Allow may still be rejected by the ward-lsm runtime.
func parse(g *grant.Grant) (*macspec.Allow, error) {
	hookArg := g.Args[0]
	if hookArg.Type != grant.ArgName {
		return nil, crex.Wrapf(ErrCompile, "expected name as hook in mac expression, got %s", hookArg)
	}
	hook := catalog().LookupHook(hookArg.Value)
	if hook == nil {
		return nil, crex.Wrapf(ErrCompile, "unknown hook %q in mac expression", hookArg.Value)
	}
	allow := &macspec.Allow{Hook: hookArg.Value}
	if g.Where != nil {
		expr, err := translateExpr(g.Where, hook)
		if err != nil {
			return nil, err
		}
		allow.Where = expr
	}
	return allow, nil
}

// Translates a grant expression tree into a MAC expression tree, validating
// field references against the hook schema.
func translateExpr(expr grant.Expr, hook *macspec.Hook) (*macspec.Expr, error) {
	switch e := expr.(type) {
	case *grant.BinaryExpr:
		return translateBinary(e, hook)
	case *grant.UnaryExpr:
		return translateUnary(e, hook)
	case *grant.CompareExpr:
		return translateCompare(e, hook)
	case *grant.InExpr:
		return translateIn(e, hook)
	case *grant.LikeExpr:
		return translateLike(e, hook)
	case *grant.BetweenExpr:
		return translateBetween(e, hook)
	case *grant.BitTestExpr:
		return translateBitTest(e, hook)
	default:
		return nil, crex.Wrapf(ErrCompile, "unknown expression type %T", expr)
	}
}

// Translates a binary boolean expression (and/or).
func translateBinary(e *grant.BinaryExpr, hook *macspec.Hook) (*macspec.Expr, error) {
	left, err := translateExpr(e.Left, hook)
	if err != nil {
		return nil, err
	}
	right, err := translateExpr(e.Right, hook)
	if err != nil {
		return nil, err
	}
	return &macspec.Expr{Type: string(e.Op), Left: left, Right: right}, nil
}

// Translates a unary boolean expression (not).
func translateUnary(e *grant.UnaryExpr, hook *macspec.Hook) (*macspec.Expr, error) {
	operand, err := translateExpr(e.Operand, hook)
	if err != nil {
		return nil, err
	}
	return &macspec.Expr{Type: "not", Operand: operand}, nil
}

// Translates a comparison expression and verifies operand type compatibility.
func translateCompare(e *grant.CompareExpr, hook *macspec.Hook) (*macspec.Expr, error) {
	lhs, err := translateOperand(e.Left, hook)
	if err != nil {
		return nil, err
	}
	rhs, err := translateOperand(e.Right, hook)
	if err != nil {
		return nil, err
	}
	if err := checkTypeCompat(e.Left, e.Right, hook); err != nil {
		return nil, err
	}
	return &macspec.Expr{Type: "cmp", Op: string(e.Op), LHS: lhs, RHS: rhs}, nil
}

// Translates an "in" set-membership expression.
func translateIn(e *grant.InExpr, hook *macspec.Hook) (*macspec.Expr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	vals := make([]*macspec.Value, 0, len(e.Values))
	for _, v := range e.Values {
		tv, err := translateOperand(v, hook)
		if err != nil {
			return nil, err
		}
		if err := checkTypeCompat(e.Field, v, hook); err != nil {
			return nil, err
		}
		vals = append(vals, tv)
	}
	return &macspec.Expr{Type: "in", Field: field, Values: vals}, nil
}

// Translates a "like" pattern match, requiring a string-typed field.
func translateLike(e *grant.LikeExpr, hook *macspec.Hook) (*macspec.Expr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	if err := requireFieldType(e.Field, hook, macspec.TypeString, "'like'"); err != nil {
		return nil, err
	}
	return &macspec.Expr{Type: "like", Field: field, Pattern: e.Pattern}, nil
}

// Translates a "between" range expression, requiring a numeric field.
func translateBetween(e *grant.BetweenExpr, hook *macspec.Hook) (*macspec.Expr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	low, err := translateOperand(e.Low, hook)
	if err != nil {
		return nil, err
	}
	high, err := translateOperand(e.High, hook)
	if err != nil {
		return nil, err
	}
	if err := requireFieldType(e.Field, hook, macspec.TypeUint64, "'between'"); err != nil {
		return nil, err
	}
	return &macspec.Expr{Type: "between", Field: field, Low: low, High: high}, nil
}

// Translates a bitwise mask test, requiring a numeric field.
func translateBitTest(e *grant.BitTestExpr, hook *macspec.Hook) (*macspec.Expr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	mask, err := translateOperand(e.Mask, hook)
	if err != nil {
		return nil, err
	}
	if err := requireFieldType(e.Field, hook, macspec.TypeUint64, "'&'"); err != nil {
		return nil, err
	}
	return &macspec.Expr{Type: "bittest", Field: field, Mask: mask}, nil
}

// Verifies that op, when it references a field, has the expected scalar type.
func requireFieldType(op grant.Operand, hook *macspec.Hook, want macspec.FieldType, label string) error {
	if !op.IsField {
		return nil
	}
	f := hook.Fields[op.Field]
	if f.Type == want {
		return nil
	}
	kind := "string field"
	if want == macspec.TypeUint64 {
		kind = "numeric field"
	}
	return crex.Wrapf(ErrCompile, "%s requires a %s, but %q is %s", label, kind, op.Field, f.Type)
}

// Translates a grant operand into a MAC value, validating field references.
func translateOperand(op grant.Operand, hook *macspec.Hook) (*macspec.Value, error) {
	if op.IsField {
		f, ok := hook.Fields[op.Field]
		if !ok {
			return nil, crex.Wrapf(ErrCompile, "field %q is not available on hook %q", op.Field, hook.Name)
		}
		if f.Sleepable && !hook.Sleepable {
			return nil, crex.Wrapf(ErrCompile, "field %q requires a sleepable hook, but %q is not sleepable", op.Field, hook.Name)
		}
		return &macspec.Value{IsField: true, Field: op.Field}, nil
	}
	switch op.Value.Type {
	case grant.ValueInt:
		return &macspec.Value{IntVal: op.Value.Int}, nil
	case grant.ValueStr:
		return &macspec.Value{StrVal: op.Value.Str}, nil
	case grant.ValueVar:
		return nil, crex.Wrapf(ErrCompile, "variable references are not supported in mac filters")
	default:
		return nil, crex.Wrapf(ErrCompile, "unsupported operand value")
	}
}

// Verifies that two operands have compatible value types.
func checkTypeCompat(left, right grant.Operand, hook *macspec.Hook) error {
	lt := resolveType(left, hook)
	rt := resolveType(right, hook)
	if lt == nil || rt == nil {
		return nil
	}
	if *lt != *rt {
		return crex.Wrapf(ErrCompile, "type mismatch between left (%s) and right (%s)", *lt, *rt)
	}
	return nil
}

// Returns the inferred field type of an operand, or nil if it cannot be
// determined.
func resolveType(op grant.Operand, hook *macspec.Hook) *macspec.FieldType {
	if op.IsField {
		f, ok := hook.Fields[op.Field]
		if !ok {
			return nil
		}
		return &f.Type
	}
	switch op.Value.Type {
	case grant.ValueInt:
		t := macspec.TypeUint64
		return &t
	case grant.ValueStr:
		t := macspec.TypeString
		return &t
	default:
		return nil
	}
}
