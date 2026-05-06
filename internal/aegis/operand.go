package aegis

// Leaf node in an expression tree (either a field reference or a literal).
//
// When IsField is true, Field holds a dotted identifier that names a runtime-
// observable property of the event being tested (e.g. "file.path", "task.uid").
// When IsField is false, Value holds a typed literal. Both LHS and RHS of
// comparisons are represented as Operand, allowing field-to-field comparisons
// such as task.uid = target.uid.
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

// Logical binary operator joining two boolean sub-expressions.
type BinaryOp string

const (
	OpAnd BinaryOp = "and" // Logical conjunction.
	OpOr  BinaryOp = "or"  // Logical disjunction.
)

// Logical unary operator.
type UnaryOp string

const (
	OpNot UnaryOp = "not" // Logical negation.
)

// Comparison operator between two operands.
type CmpOp string

const (
	CmpEq  CmpOp = "="  // Equal.
	CmpNeq CmpOp = "!=" // Not equal.
	CmpGt  CmpOp = ">"  // Greater than.
	CmpGte CmpOp = ">=" // Greater than or equal.
	CmpLt  CmpOp = "<"  // Less than.
	CmpLte CmpOp = "<=" // Less than or equal.
)
