package macspec

// Normalised expression node in the MAC subsystem model.
//
// Type selects the expression variant and determines which fields are
// populated. For binary nodes ("and", "or") Left and Right are set;
// for unary ("not") only Operand; for comparisons ("cmp") Op, LHS, and
// RHS; and so on for "in", "like", "between", and "bittest".
type Expr struct {
	Type    string   `json:"type"`    // Expression type discriminator.
	Left    *Expr    `json:"left"`    // Left child for binary expressions.
	Right   *Expr    `json:"right"`   // Right child for binary expressions.
	Operand *Expr    `json:"operand"` // Operand for unary expressions.
	Op      string   `json:"op"`      // Operator for comparison expressions.
	LHS     *Value   `json:"lhs"`     // Left-hand side value for comparison expressions.
	RHS     *Value   `json:"rhs"`     // Right-hand side value for comparison expressions.
	Field   *Value   `json:"field"`   // Field value for field-based expressions.
	Values  []*Value `json:"values"`  // Slice of values for multi-value expressions.
	Pattern string   `json:"pattern"` // Pattern for pattern-matching expressions.
	Low     *Value   `json:"low"`     // Low value for between expressions.
	High    *Value   `json:"high"`    // High value for between expressions.
	Mask    *Value   `json:"mask"`    // Mask value for bit-test expressions.
}

// Typed value representing a field reference or literal.
type Value struct {
	IsField bool   `json:"is_field"` // Whether this value is a field reference (true) or a literal (false).
	Field   string `json:"field"`    // Field name when IsField is true.
	IntVal  uint64 `json:"int_val"`  // Integer literal value when IsField is false.
	StrVal  string `json:"str_val"`  // String literal value when IsField is false.
}

// Whether two expression trees are structurally identical.
func exprEqual(a, b *Expr) bool {
	if a == nil || b == nil {
		return a == b
	}
	return scalarFieldsEqual(a, b) &&
		childrenEqual(a, b) &&
		valuesEqual(a, b)
}

// Compares the scalar (non-pointer, non-slice) discriminator fields.
func scalarFieldsEqual(a, b *Expr) bool {
	return a.Type == b.Type && a.Op == b.Op && a.Pattern == b.Pattern
}

// Compares the recursively-typed Expr children.
func childrenEqual(a, b *Expr) bool {
	return exprEqual(a.Left, b.Left) &&
		exprEqual(a.Right, b.Right) &&
		exprEqual(a.Operand, b.Operand)
}

// Compares the Value-typed operand fields and the Values slice.
func valuesEqual(a, b *Expr) bool {
	if !valueEqual(a.LHS, b.LHS) || !valueEqual(a.RHS, b.RHS) {
		return false
	}
	if !valueEqual(a.Field, b.Field) {
		return false
	}
	if !valueEqual(a.Low, b.Low) || !valueEqual(a.High, b.High) {
		return false
	}
	if !valueEqual(a.Mask, b.Mask) {
		return false
	}
	if len(a.Values) != len(b.Values) {
		return false
	}
	for i := range a.Values {
		if !valueEqual(a.Values[i], b.Values[i]) {
			return false
		}
	}
	return true
}

// Whether two values are identical.
func valueEqual(a, b *Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.IsField == b.IsField &&
		a.Field == b.Field &&
		a.IntVal == b.IntVal &&
		a.StrVal == b.StrVal
}

// Returns a deep copy of e, or nil when e is nil.
func cloneExpr(e *Expr) *Expr {
	if e == nil {
		return nil
	}
	copy := &Expr{
		Type:    e.Type,
		Left:    cloneExpr(e.Left),
		Right:   cloneExpr(e.Right),
		Operand: cloneExpr(e.Operand),
		Op:      e.Op,
		LHS:     cloneValue(e.LHS),
		RHS:     cloneValue(e.RHS),
		Field:   cloneValue(e.Field),
		Pattern: e.Pattern,
		Low:     cloneValue(e.Low),
		High:    cloneValue(e.High),
		Mask:    cloneValue(e.Mask),
	}
	if len(e.Values) != 0 {
		copy.Values = make([]*Value, 0, len(e.Values))
		for _, value := range e.Values {
			copy.Values = append(copy.Values, cloneValue(value))
		}
	}
	return copy
}

// Returns a shallow copy of v, or nil when v is nil.
func cloneValue(v *Value) *Value {
	if v == nil {
		return nil
	}
	copy := *v
	return &copy
}
