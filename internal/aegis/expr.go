package aegis

import (
	"fmt"
	"strings"
)

// Node in a where-clause expression tree.
//
// The interface is intentionally open: any type that renders to a canonical
// source form via fmt.Stringer satisfies it. Subsystems may introduce their
// own Expr nodes if they need to extend the grammar. The seven node types
// produced by Parse are BinaryExpr, UnaryExpr, CompareExpr, InExpr, LikeExpr,
// BetweenExpr, and BitTestExpr.
type Expr interface {
	fmt.Stringer
}

// Logical conjunction or disjunction of two sub-expressions.
//
// Produced by parsing "a and b" or "a or b". Left-associative; nested
// expressions of the same operator chain to the left.
type BinaryExpr struct {
	Op    BinaryOp // Operator joining the two sides.
	Left  Expr     // Left operand.
	Right Expr     // Right operand.
}

// Renders the expression as "(left op right)".
//
// Outer parentheses are always emitted to make precedence unambiguous, even
// when the original source omitted them.
func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Op, e.Right)
}

// Logical negation of a sub-expression.
//
// Produced by parsing "not expr". Note that "not like" is parsed into a
// LikeExpr with Negated set, not into a UnaryExpr wrapping a LikeExpr.
type UnaryExpr struct {
	Op      UnaryOp // Operator applied to the operand.
	Operand Expr    // Sub-expression being negated.
}

// Renders the expression as "op operand".
func (e *UnaryExpr) String() string {
	return fmt.Sprintf("%s %s", e.Op, e.Operand)
}

// Comparison between two operands using one of the six relational operators.
//
// Both sides are Operand and may independently be field references or literals,
// allowing field-to-field comparisons (such as task.uid = target.uid).
type CompareExpr struct {
	Left  Operand // Left-hand side.
	Op    CmpOp   // Comparison operator.
	Right Operand // Right-hand side.
}

// Renders the expression as "left op right".
func (e *CompareExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.Left, e.Op, e.Right)
}

// Set-membership test on a field against a list of values.
//
// Produced by parsing "field in (v1, v2, ...)". The value list is preserved
// in source order; deduplication and semantic interpretation are deferred to
// the subsystem.
type InExpr struct {
	Field  Operand   // Field being tested.
	Values []Operand // Candidate values, in source order.
}

// Renders the expression as "field in (v1, v2, ...)".
func (e *InExpr) String() string {
	parts := make([]string, len(e.Values))
	for i, v := range e.Values {
		parts[i] = v.String()
	}
	return fmt.Sprintf("%s in (%s)", e.Field, strings.Join(parts, ", "))
}

// Pattern-match test on a field against a string pattern.
//
// Produced by parsing "field like STRING" or "field not like STRING". The
// pattern is preserved verbatim after escape resolution; its glob syntax is
// interpreted by the subsystem at evaluation time, not by the parser.
type LikeExpr struct {
	Field   Operand // Field being matched.
	Pattern string  // Glob pattern, with source-level escapes resolved.
	Negated bool    // True when parsed from "not like".
}

// Renders the expression as "field like \"pat\"" or "field not like \"pat\"".
func (e *LikeExpr) String() string {
	op := "like"
	if e.Negated {
		op = "not like"
	}
	return fmt.Sprintf("%s %s %q", e.Field, op, e.Pattern)
}

// Inclusive range test on a field.
//
// Produced by parsing "field between low and high". The bounds are operands
// so they may be field references or literals; their ordering and types are
// not validated by the parser.
type BetweenExpr struct {
	Field Operand // Field being tested.
	Low   Operand // Inclusive lower bound.
	High  Operand // Inclusive upper bound.
}

// Renders the expression as "field between low and high".
func (e *BetweenExpr) String() string {
	return fmt.Sprintf("%s between %s and %s", e.Field, e.Low, e.High)
}

// Bitwise mask test on a field.
//
// Produced by parsing "field & mask" (truth test, true when any bit set) or
// "field & mask = value" (equality test on the masked bits). When Val is nil
// the expression is the truth-test form.
type BitTestExpr struct {
	Field Operand  // Field being tested.
	Mask  Operand  // Mask applied to the field.
	Val   *Operand // Optional expected value of the masked bits; nil for the truth-test form.
}

// Renders the expression as "field & mask" or "field & mask = val".
func (e *BitTestExpr) String() string {
	if e.Val == nil {
		return fmt.Sprintf("%s & %s", e.Field, e.Mask)
	}
	return fmt.Sprintf("%s & %s = %s", e.Field, e.Mask, *e.Val)
}
