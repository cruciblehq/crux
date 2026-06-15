package mac

import (
	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/crex"
)

// Normalised expression node in the MAC subsystem model.
//
// Type selects the expression variant and determines which fields are populated.
// For binary nodes ("and", "or") Left and Right are set; for unary ("not") only
// Operand; for comparisons ("cmp") Op, LHS, and RHS; and so on for "in", "like",
// "between", and "bittest".
type MACExpr struct {
	Type    string      `codec:"type"`    // Expression type discriminator.
	Left    *MACExpr    `codec:"left"`    // Left child for binary expressions.
	Right   *MACExpr    `codec:"right"`   // Right child for binary expressions.
	Operand *MACExpr    `codec:"operand"` // Operand for unary expressions.
	Op      string      `codec:"op"`      // Comparison operator.
	LHS     *MACValue   `codec:"lhs"`     // Left-hand side value for comparison expressions.
	RHS     *MACValue   `codec:"rhs"`     // Right-hand side value for comparison expressions.
	Field   *MACValue   `codec:"field"`   // Field reference for field-based expressions.
	Values  []*MACValue `codec:"values"`  // Values for multi-value expressions such as "in".
	Pattern string      `codec:"pattern"` // Pattern for pattern-matching expressions.
	Low     *MACValue   `codec:"low"`     // Lower bound for between expressions.
	High    *MACValue   `codec:"high"`    // Upper bound for between expressions.
	Mask    *MACValue   `codec:"mask"`    // Mask value for bit-test expressions.
}

// Expression type discriminators for the Type field of a MACExpr.
//
// The boolean operators reuse the AGL operator constants; the remaining
// discriminators are specific to the normalised MAC model.
const (
	exprAnd     = string(agl.OpAnd) // Logical conjunction of Left and Right.
	exprOr      = string(agl.OpOr)  // Logical disjunction of Left and Right.
	exprNot     = string(agl.OpNot) // Logical negation of Operand.
	exprCmp     = "cmp"             // Comparison of LHS and RHS with Op.
	exprIn      = "in"              // Membership test of Field against Values.
	exprLike    = "like"            // Pattern match of Field against Pattern.
	exprBetween = "between"         // Range test of Field within Low and High.
	exprBitTest = "bittest"         // Bit-mask test of Field against Mask.
)

// Deep-clones a MACExpr.
func cloneMACExpr(e *MACExpr) *MACExpr {
	if e == nil {
		return nil
	}
	vals := make([]*MACValue, len(e.Values))
	for i, v := range e.Values {
		vals[i] = cloneMACValue(v)
	}
	return &MACExpr{
		Type:    e.Type,
		Left:    cloneMACExpr(e.Left),
		Right:   cloneMACExpr(e.Right),
		Operand: cloneMACExpr(e.Operand),
		Op:      e.Op,
		LHS:     cloneMACValue(e.LHS),
		RHS:     cloneMACValue(e.RHS),
		Field:   cloneMACValue(e.Field),
		Values:  vals,
		Pattern: e.Pattern,
		Low:     cloneMACValue(e.Low),
		High:    cloneMACValue(e.High),
		Mask:    cloneMACValue(e.Mask),
	}
}

// Whether two MACExpr trees are structurally equal.
func macExprEqual(a, b *MACExpr) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type != b.Type || a.Op != b.Op || a.Pattern != b.Pattern {
		return false
	}
	if !macValueEqual(a.LHS, b.LHS) || !macValueEqual(a.RHS, b.RHS) {
		return false
	}
	if !macValueEqual(a.Field, b.Field) || !macValueEqual(a.Mask, b.Mask) {
		return false
	}
	if !macValueEqual(a.Low, b.Low) || !macValueEqual(a.High, b.High) {
		return false
	}
	return macExprEqual(a.Left, b.Left) &&
		macExprEqual(a.Right, b.Right) &&
		macExprEqual(a.Operand, b.Operand) &&
		macValuesEqual(a.Values, b.Values)
}

// Validates the expression tree node.
//
// Type must be one of the known expression types and the required operands for
// that type must be present. Child nodes are validated recursively.
func (e *MACExpr) Validate() error {
	switch e.Type {
	case exprAnd, exprOr:
		return e.validateBinary()
	case exprNot:
		return e.validateUnary()
	case exprCmp:
		return e.validateComparison()
	case exprIn:
		return e.validateIn()
	case exprLike:
		return e.validateLike()
	case exprBetween:
		return e.validateBetween()
	case exprBitTest:
		return e.validateBitTest()
	default:
		return crex.Wrapf(ErrInvalidMACExpr, "unknown type %q", e.Type)
	}
}

// Validates each provided child expression node, wrapping any failure.
func validateExprs(children ...*MACExpr) error {
	for _, c := range children {
		if err := c.Validate(); err != nil {
			return crex.Wrap(ErrInvalidMACExpr, err)
		}
	}
	return nil
}

// Validates each provided value, wrapping any failure.
func validateValues(values ...*MACValue) error {
	for _, v := range values {
		if err := v.Validate(); err != nil {
			return crex.Wrap(ErrInvalidMACExpr, err)
		}
	}
	return nil
}

// Validates a binary node, requiring both child operands.
func (e *MACExpr) validateBinary() error {
	if e.Left == nil || e.Right == nil {
		return crex.Wrapf(ErrInvalidMACExpr, "%q requires left and right operands", e.Type)
	}
	return validateExprs(e.Left, e.Right)
}

// Validates a unary node, requiring its single operand.
func (e *MACExpr) validateUnary() error {
	if e.Operand == nil {
		return crex.Wrapf(ErrInvalidMACExpr, "%q requires an operand", e.Type)
	}
	return validateExprs(e.Operand)
}

// Validates a comparison node, requiring an operator and both sides.
func (e *MACExpr) validateComparison() error {
	if e.Op == "" || e.LHS == nil || e.RHS == nil {
		return crex.Wrapf(ErrInvalidMACExpr, "%q requires op, lhs, and rhs", e.Type)
	}
	return validateValues(e.LHS, e.RHS)
}

// Validates a membership node, requiring a field and non-nil values.
func (e *MACExpr) validateIn() error {
	if e.Field == nil || len(e.Values) == 0 {
		return crex.Wrapf(ErrInvalidMACExpr, "%q requires a field and at least one value", e.Type)
	}
	if err := validateValues(e.Field); err != nil {
		return err
	}
	for i, v := range e.Values {
		if v == nil {
			return crex.Wrapf(ErrInvalidMACExpr, "%q has nil value at index %d", e.Type, i)
		}
		if err := validateValues(v); err != nil {
			return err
		}
	}
	return nil
}

// Validates a pattern-match node, requiring a field and pattern.
func (e *MACExpr) validateLike() error {
	if e.Field == nil || e.Pattern == "" {
		return crex.Wrapf(ErrInvalidMACExpr, "%q requires a field and pattern", e.Type)
	}
	return validateValues(e.Field)
}

// Validates a range node, requiring a field and both bounds.
func (e *MACExpr) validateBetween() error {
	if e.Field == nil || e.Low == nil || e.High == nil {
		return crex.Wrapf(ErrInvalidMACExpr, "%q requires field, low, and high", e.Type)
	}
	return validateValues(e.Field, e.Low, e.High)
}

// Validates a bit-test node, requiring a field and mask.
func (e *MACExpr) validateBitTest() error {
	if e.Field == nil || e.Mask == nil {
		return crex.Wrapf(ErrInvalidMACExpr, "%q requires field and mask", e.Type)
	}
	return validateValues(e.Field, e.Mask)
}
