package aegis

import (
	"fmt"
	"strings"
)

// Lexical category of an Arg or Kwarg as it appeared in the source.
//
// Preserved on the AST so downstream consumers can tell symbol types apart (a
// name from an integer, ASCII from Unicode, etc) without re-parsing. Subsystems
// re-parse the Value text as needed to obtain a typed integer or to distinguish
// a glob from a quoted string.
type ArgType int

const (
	ArgName       ArgType = iota // Name token (identifier, dotted name).
	ArgInt                       // Unsigned integer literal in a C-style base.
	ArgQuantity                  // Unsigned integer literal with a unit suffix (e.g. 1Gi, 500m).
	ArgStrASCII                  // Quoted ASCII string literal or unquoted glob.
	ArgStrUnicode                // Quoted Unicode string literal (u"..." prefix).
	ArgVar                       // Variable reference; Value holds the name without the leading '$'.
)

// Single positional argument or keyword-argument value.
//
// Type names the lexical category the argument was written as. Value holds
// the decoded text: for string types escape sequences (\" and \\) have been
// resolved and surrounding quotes stripped; for names and integers it is
// the verbatim source text. Subsystems re-parse Value when they need a
// typed integer.
type Arg struct {
	Type  ArgType // Lexical category.
	Value string  // Decoded text content of the argument.
}

// Renders the argument in canonical source form.
//
// Strings are re-quoted with the standard escape rules; Unicode strings carry
// the u"..." prefix; variables are rendered with the leading '$'; names and
// integers are emitted verbatim.
func (a Arg) String() string {
	switch a.Type {
	case ArgName, ArgInt, ArgQuantity:
		return a.Value
	case ArgStrASCII:
		return fmt.Sprintf("%q", a.Value)
	case ArgStrUnicode:
		return fmt.Sprintf("u%q", a.Value)
	case ArgVar:
		return "$" + a.Value
	default:
		return "<unknown>"
	}
}

// A single key=value keyword argument.
//
// Key is a NAME (possibly dotted, e.g. "cpu.weight"). Value carries the type
// and decoded text of the right-hand-side scalar. Kwargs only appear after
// all positional arguments within a grant.
type Kwarg struct {
	Key   string // Verbatim NAME on the left-hand side.
	Value Arg    // Right-hand-side scalar.
}

// Renders the kwarg in canonical "key=value" form.
func (k Kwarg) String() string {
	return k.Key + "=" + k.Value.String()
}

// Leaf node in an expression tree (either a field reference or a literal).
//
// When IsField is true, Field holds a dotted identifier that names a
// runtime-observable property of the event being tested (e.g. "file.path",
// "arg0", "task.uid"). When IsField is false, Value holds a typed literal.
// Both LHS and RHS of comparisons are represented as Operand, allowing
// field-to-field comparisons such as task.uid = target.uid.
type Operand struct {
	IsField bool   // True when this operand is a field reference rather than a literal.
	Field   string // Dotted field name. Valid only when IsField is true.
	Value   Value  // Literal value. Valid only when IsField is false.
}

// Renders the operand in canonical source form.
//
// Field references are rendered as their dotted name; literals delegate to
// Value.String.
func (o Operand) String() string {
	if o.IsField {
		return o.Field
	}
	return o.Value.String()
}

// Discriminates the type stored in a Value.
type ValueType int

const (
	ValueNone ValueType = iota // No value; uninitialised sentinel.
	ValueInt                   // Unsigned integer literal (decimal or 0x hex).
	ValueStr                   // String literal with escape sequences resolved.
	ValueVar                   // Variable reference; Str holds the name without the leading '$'.
)

// Source-text encoding of a string literal.
//
// ASCII strings are written as quoted "..." literals or as unquoted globs.
// Unicode strings are written with the u"..." prefix. The encoding is
// preserved on the AST so that downstream consumers (semantic checkers,
// rule serialisers) can distinguish between the two without re-parsing.
type StrEncoding int

const (
	StrASCII   StrEncoding = iota // Quoted ASCII literal or unquoted glob.
	StrUnicode                    // Quoted Unicode literal (u"..." prefix).
)

// Typed literal value carried by a non-field Operand.
//
// Exactly one of Int or Str is meaningful depending on Type. When Type
// is ValueNone neither field is populated and the Operand should not be
// inspected. StrEncoding is meaningful only when Type is ValueStr.
type Value struct {
	Type        ValueType   // Discriminator indicating which field holds the data.
	Int         uint64      // Unsigned 64-bit integer. Valid when Type is ValueInt.
	Str         string      // Resolved string content. Valid when Type is ValueStr.
	StrEncoding StrEncoding // Source-text encoding of the string literal. Valid when Type is ValueStr.
}

// Renders the value in canonical source form.
//
// Integers are emitted in decimal regardless of their original base; strings
// are re-quoted with the standard escape rules and prefixed with 'u' when the
// encoding is Unicode; variables are rendered with the leading '$'. The
// uninitialised ValueNone case is rendered as "<none>" and indicates a bug
// in the producer rather than a legitimate source form.
func (v Value) String() string {
	switch v.Type {
	case ValueInt:
		return fmt.Sprintf("%d", v.Int)
	case ValueStr:
		if v.StrEncoding == StrUnicode {
			return fmt.Sprintf("u%q", v.Str)
		}
		return fmt.Sprintf("%q", v.Str)
	case ValueVar:
		return "$" + v.Str
	default:
		return "<none>"
	}
}

// Logical binary operator joining two boolean sub-expressions.
//
// The underlying string is the canonical source spelling, so a BinaryOp can
// be formatted directly with %s and round-trips through Parse.
type BinaryOp string

const (
	OpAnd BinaryOp = "and" // Logical conjunction.
	OpOr  BinaryOp = "or"  // Logical disjunction.
)

// Logical unary operator applied to a single boolean sub-expression.
//
// The underlying string is the canonical source spelling, so a UnaryOp can be
// formatted directly with %s and round-trips through Parse.
type UnaryOp string

const (
	OpNot UnaryOp = "not" // Logical negation.
)

// Comparison operator between two operands.
//
// The underlying string is the canonical source spelling, so a CmpOp can be
// formatted directly with %s and round-trips through Parse.
type CmpOp string

const (
	CmpEq  CmpOp = "="  // Equal.
	CmpNeq CmpOp = "!=" // Not equal.
	CmpGt  CmpOp = ">"  // Greater than.
	CmpGte CmpOp = ">=" // Greater than or equal.
	CmpLt  CmpOp = "<"  // Less than.
	CmpLte CmpOp = "<=" // Less than or equal.
)

// Node in a where-clause expression tree.
//
// The interface is intentionally open. Any type that renders to a canonical
// source form via fmt.Stringer satisfies it; subsystems may introduce their
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
// Both sides are Operand and may independently be field references or
// literals, allowing field-to-field comparisons such as task.uid = target.uid.
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
