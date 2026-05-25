package mac

import (
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/spec"

	"github.com/cruciblehq/crux/resource/affordance/agl"
	"github.com/cruciblehq/crux/resource/affordance/subsystem"
)

// Implementation for MAC (LSM) grants.
type Subsystem struct {
	spec *spec.MAC // Pointer to the unified spec's mac section.
}

// Returns a Subsystem wired to mutate spec.
func New(spec *spec.MAC) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the mac subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameMAC
}

// Returns the deduplication key for a mac grant.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) == 0 {
		return ""
	}
	return g.Args[0].Value
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
func parse(g *agl.Model) (*spec.MACAllow, error) {
	hookArg := g.Args[0]
	if hookArg.Type != agl.ArgName {
		return nil, crex.Wrapf(ErrCompile, "expected name as hook in mac expression")
	}
	hook := catalog().LookupHook(hookArg.Value)
	if hook == nil {
		return nil, crex.Wrapf(ErrCompile, "unknown hook %q in mac expression", hookArg.Value)
	}
	allow := &spec.MACAllow{Hook: hookArg.Value}
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
func translateExpr(expr agl.Expr, hook *Hook) (*spec.MACExpr, error) {
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
		return nil, crex.Wrapf(ErrCompile, "unknown expression type %T", expr)
	}
}

// Translates a binary boolean expression (and/or).
func translateBinary(e *agl.BinaryExpr, hook *Hook) (*spec.MACExpr, error) {
	left, err := translateExpr(e.Left, hook)
	if err != nil {
		return nil, err
	}
	right, err := translateExpr(e.Right, hook)
	if err != nil {
		return nil, err
	}
	return &spec.MACExpr{Type: string(e.Op), Left: left, Right: right}, nil
}

// Translates a unary boolean expression (not).
func translateUnary(e *agl.UnaryExpr, hook *Hook) (*spec.MACExpr, error) {
	operand, err := translateExpr(e.Operand, hook)
	if err != nil {
		return nil, err
	}
	return &spec.MACExpr{Type: "not", Operand: operand}, nil
}

// Translates a comparison expression and verifies operand type compatibility.
func translateCompare(e *agl.CompareExpr, hook *Hook) (*spec.MACExpr, error) {
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
	return &spec.MACExpr{Type: "cmp", Op: string(e.Op), LHS: lhs, RHS: rhs}, nil
}

// Translates an "in" set-membership expression.
func translateIn(e *agl.InExpr, hook *Hook) (*spec.MACExpr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	vals := make([]*spec.MACValue, 0, len(e.Values))
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
	return &spec.MACExpr{Type: "in", Field: field, Values: vals}, nil
}

// Translates a "like" pattern match, requiring a string-typed field.
func translateLike(e *agl.LikeExpr, hook *Hook) (*spec.MACExpr, error) {
	field, err := translateOperand(e.Field, hook)
	if err != nil {
		return nil, err
	}
	if err := requireFieldType(e.Field, hook, TypeString, "'like'"); err != nil {
		return nil, err
	}
	return &spec.MACExpr{Type: "like", Field: field, Pattern: e.Pattern}, nil
}

// Translates a "between" range expression, requiring a numeric field.
func translateBetween(e *agl.BetweenExpr, hook *Hook) (*spec.MACExpr, error) {
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
	return &spec.MACExpr{Type: "between", Field: field, Low: low, High: high}, nil
}

// Translates a bitwise mask test, requiring a numeric field.
func translateBitTest(e *agl.BitTestExpr, hook *Hook) (*spec.MACExpr, error) {
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
	return &spec.MACExpr{Type: "bittest", Field: field, Mask: mask}, nil
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
	kind := "string field"
	if want == TypeUint64 {
		kind = "numeric field"
	}
	return crex.Wrapf(ErrCompile, "%s requires a %s, but %q is %s", label, kind, op.Field, f.Type)
}

// Translates a grant operand into a MAC value, validating field references.
func translateOperand(op agl.Operand, hook *Hook) (*spec.MACValue, error) {
	if op.IsField {
		f, ok := hook.Fields[op.Field]
		if !ok {
			return nil, crex.Wrapf(ErrCompile, "field %q is not available on hook %q", op.Field, hook.Name)
		}
		if f.Sleepable && !hook.Sleepable {
			return nil, crex.Wrapf(ErrCompile, "field %q requires a sleepable hook, but %q is not sleepable", op.Field, hook.Name)
		}
		return &spec.MACValue{IsField: true, Field: op.Field}, nil
	}
	switch op.Value.Type {
	case agl.ValueInt:
		return &spec.MACValue{IntVal: op.Value.Int}, nil
	case agl.ValueStr:
		return &spec.MACValue{StrVal: op.Value.Str}, nil
	case agl.ValueVar:
		return nil, crex.Wrapf(ErrCompile, "variable references are not supported in mac filters")
	default:
		return nil, crex.Wrapf(ErrCompile, "unsupported operand value")
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
		return crex.Wrapf(ErrCompile, "type mismatch between left (%s) and right (%s)", *lt, *rt)
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
