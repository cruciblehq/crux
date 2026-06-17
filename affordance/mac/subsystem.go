package mac

import (
	"github.com/cruciblehq/crux/crex"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

// Implementation for MAC (LSM) grants.
type Subsystem struct {
	spec *Spec // Pointer to the unified spec's mac section.
}

// Returns a Subsystem wired to mutate spec.
func New(spec *Spec) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the mac subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameMAC
}

// Applies a parsed grant to the wired-in section.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	allow, err := parse(g)
	if err != nil {
		return err
	}
	s.spec.Apply(allow)
	return nil
}

// Validates the grant's structural shape against what the mac subsystem accepts.
func check(g *agl.Model) error {
	if len(g.Kwargs) != 0 {
		return crex.Newf(ErrCompile, "unexpected keyword arguments in mac expression")
	}
	if len(g.Args) != 1 {
		return crex.Newf(ErrCompile, "wrong number of arguments in mac expression")
	}
	return nil
}

// Extracts and validates the grant's content into an Allow rule.
//
// Validates that the hook exists and that the where clause, if present, is
// well-formed and references only fields available on the hook. Does not check
// for semantic validity of the where clause beyond field reference checks;
// the resulting Allow may still be rejected by the ward-lsm runtime.
func parse(g *agl.Model) (*MACAllow, error) {
	hookArg := g.Args[0]
	if hookArg.Type != agl.ArgName {
		return nil, crex.Newf(ErrCompile, "expected name as hook in mac expression")
	}
	hook := catalog().LookupHook(hookArg.Value)
	if hook == nil {
		return nil, crex.Newf(ErrCompile, "unknown hook %q in mac expression", hookArg.Value)
	}
	allow := &MACAllow{Hook: hookArg.Value}
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
func translateExpr(expr agl.Expr, hook *Hook) (*MACExpr, error) {
	switch e := expr.(type) {
	case *agl.BinaryExpr:
		return translateBinary(e, hook)
	case *agl.UnaryExpr:
		return translateUnary(e, hook)
	case *agl.CompareExpr:
		return translateCompare(e, hook)
	case *agl.InExpr:
		return translateIn(e, hook)
	case *agl.LikeExpr:
		return translateLike(e, hook)
	case *agl.BetweenExpr:
		return translateBetween(e, hook)
	case *agl.BitTestExpr:
		return translateBitTest(e, hook)
	default:
		return nil, crex.Newf(ErrCompile, "unknown expression type %T", expr)
	}
}

// Translates a binary boolean expression (and/or).
func translateBinary(e *agl.BinaryExpr, hook *Hook) (*MACExpr, error) {
	left, err := translateExpr(e.Left, hook)
	if err != nil {
		return nil, err
	}
	right, err := translateExpr(e.Right, hook)
	if err != nil {
		return nil, err
	}
	return &MACExpr{Type: string(e.Op), Left: left, Right: right}, nil
}

// Translates a unary boolean expression (not).
func translateUnary(e *agl.UnaryExpr, hook *Hook) (*MACExpr, error) {
	operand, err := translateExpr(e.Operand, hook)
	if err != nil {
		return nil, err
	}
	return &MACExpr{Type: exprNot, Operand: operand}, nil
}

// Translates a comparison expression and verifies operand type compatibility.
func translateCompare(e *agl.CompareExpr, hook *Hook) (*MACExpr, error) {
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
	return &MACExpr{Type: exprCmp, Op: string(e.Op), LHS: lhs, RHS: rhs}, nil
}

// Translates an "in" set-membership expression.
func translateIn(e *agl.InExpr, hook *Hook) (*MACExpr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	vals := make([]*MACValue, 0, len(e.Values))
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
	return &MACExpr{Type: exprIn, Field: field, Values: vals}, nil
}

// Translates a "like" pattern match, requiring a string-typed field.
func translateLike(e *agl.LikeExpr, hook *Hook) (*MACExpr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	if err := requireFieldType(e.Field, hook, TypeString, "'like'"); err != nil {
		return nil, err
	}
	return &MACExpr{Type: exprLike, Field: field, Pattern: e.Pattern}, nil
}

// Translates a "between" range expression, requiring a numeric field.
func translateBetween(e *agl.BetweenExpr, hook *Hook) (*MACExpr, error) {
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
	if err := requireFieldType(e.Field, hook, TypeUint64, "'between'"); err != nil {
		return nil, err
	}
	return &MACExpr{Type: exprBetween, Field: field, Low: low, High: high}, nil
}

// Translates a bitwise mask test, requiring a numeric field.
func translateBitTest(e *agl.BitTestExpr, hook *Hook) (*MACExpr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	mask, err := translateOperand(e.Mask, hook)
	if err != nil {
		return nil, err
	}
	if err := requireFieldType(e.Field, hook, TypeUint64, "'&'"); err != nil {
		return nil, err
	}
	return &MACExpr{Type: exprBitTest, Field: field, Mask: mask}, nil
}

// Verifies that op, when it references a field, has the expected scalar type.
func requireFieldType(op agl.Operand, hook *Hook, want FieldType, label string) error {
	if !op.IsField {
		return nil
	}
	f := hook.Fields[op.Field]
	if f.Type == want {
		return nil
	}
	fieldType := "string field"
	if want == TypeUint64 {
		fieldType = "numeric field"
	}
	return crex.Newf(ErrCompile, "%s requires a %s, but %q is %s", label, fieldType, op.Field, f.Type)
}

// Translates a grant operand into a MAC value, validating field references.
func translateOperand(op agl.Operand, hook *Hook) (*MACValue, error) {
	if op.IsField {
		f, ok := hook.Fields[op.Field]
		if !ok {
			return nil, crex.Newf(ErrCompile, "field %q is not available on hook %q", op.Field, hook.Name)
		}
		if f.Sleepable && !hook.Sleepable {
			return nil, crex.Newf(ErrCompile, "field %q requires a sleepable hook, but %q is not sleepable", op.Field, hook.Name)
		}
		return &MACValue{IsField: true, Field: op.Field}, nil
	}
	switch op.Value.Type {
	case agl.ValueInt:
		return &MACValue{IntVal: op.Value.Int}, nil
	case agl.ValueStr:
		return &MACValue{StrVal: op.Value.Str}, nil
	case agl.ValueVar:
		return nil, crex.Newf(ErrCompile, "variable references are not supported in mac filters")
	default:
		return nil, crex.Newf(ErrCompile, "unsupported operand value")
	}
}

// Verifies that two operands have compatible value types.
func checkTypeCompat(left, right agl.Operand, hook *Hook) error {
	lt := resolveType(left, hook)
	rt := resolveType(right, hook)
	if lt == nil || rt == nil {
		return nil
	}
	if *lt != *rt {
		return crex.Newf(ErrCompile, "type mismatch between left (%s) and right (%s)", *lt, *rt)
	}
	return nil
}

// Returns the inferred field type of an operand, or nil if it cannot be
// determined.
func resolveType(op agl.Operand, hook *Hook) *FieldType {
	if op.IsField {
		f, ok := hook.Fields[op.Field]
		if !ok {
			return nil
		}
		return &f.Type
	}
	switch op.Value.Type {
	case agl.ValueInt:
		t := TypeUint64
		return &t
	case agl.ValueStr:
		t := TypeString
		return &t
	default:
		return nil
	}
}
