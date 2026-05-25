package spec

// Accumulated MAC grant spec.
//
// Built up by successive calls to Apply during affordance evaluation.
// Duplicate rules are suppressed so the slice contains only the minimal set
// of distinct hook allows required by the workload.
type MAC struct {
	// Granted LSM hook allow rules.
	Rules []*MACAllow `codec:"rules"`
}

// Single LSM hook allow rule produced by a MAC grant.
//
// Hook names the kernel LSM callback point at which the rule is evaluated.
// Where constrains the allow to invocations matching the predicate tree; a
// nil Where makes the rule unconditional.
type MACAllow struct {
	// Kernel LSM hook name.
	Hook string `codec:"hook"`
	// Where-clause expression tree.
	//
	// A nil value makes the allow unconditional.
	Where *MACExpr `codec:"where"`
}

// Normalised expression node in the MAC subsystem model.
//
// Type selects the expression variant and determines which fields are
// populated. For binary nodes ("and", "or") Left and Right are set; for
// unary ("not") only Operand; for comparisons ("cmp") Op, LHS, and RHS;
// and so on for "in", "like", "between", and "bittest".
type MACExpr struct {
	// Expression type discriminator.
	Type string `codec:"type"`
	// Left child for binary expressions.
	Left *MACExpr `codec:"left"`
	// Right child for binary expressions.
	Right *MACExpr `codec:"right"`
	// Operand for unary expressions.
	Operand *MACExpr `codec:"operand"`
	// Comparison operator.
	Op string `codec:"op"`
	// Left-hand side value for comparison expressions.
	LHS *MACValue `codec:"lhs"`
	// Right-hand side value for comparison expressions.
	RHS *MACValue `codec:"rhs"`
	// Field reference for field-based expressions.
	Field *MACValue `codec:"field"`
	// Values for multi-value expressions such as "in".
	Values []*MACValue `codec:"values"`
	// Pattern for pattern-matching expressions.
	Pattern string `codec:"pattern"`
	// Lower bound for between expressions.
	Low *MACValue `codec:"low"`
	// Upper bound for between expressions.
	High *MACValue `codec:"high"`
	// Mask value for bit-test expressions.
	Mask *MACValue `codec:"mask"`
}

// Typed value representing either a kernel field reference or a literal.
//
// When IsField is true, Field holds the kernel field name and the integer and
// string fields are unused. When IsField is false, either IntVal or StrVal
// carries the literal depending on the expression context.
type MACValue struct {
	// Marks this value as a field reference rather than a literal.
	IsField bool `codec:"is_field"`
	// Kernel field name.
	//
	// Meaningful only when IsField is true.
	Field string `codec:"field"`
	// Integer literal value.
	//
	// Meaningful only when IsField is false.
	IntVal uint64 `codec:"int_val"`
	// String literal value.
	//
	// Meaningful only when IsField is false.
	StrVal string `codec:"str_val"`
}

// Appends allow to the spec's rule list if it is not already present.
//
// Returns true if the rule was added (false means a rule with the same hook
// and where clause already existed, and the call was a no-op).
func (s *MAC) Apply(r *MACAllow) bool {
	for _, existing := range s.Rules {
		if existing.Hook == r.Hook && macExprEqual(existing.Where, r.Where) {
			return false
		}
	}
	s.Rules = append(s.Rules, cloneMACAllow(r))
	return true
}

// Deep-clones a MACAllow.
func cloneMACAllow(a *MACAllow) *MACAllow {
	return &MACAllow{
		Hook:  a.Hook,
		Where: cloneMACExpr(a.Where),
	}
}

// Deep-clones a MACExpr, returning nil for a nil input.
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

// Deep-clones a MACValue, returning nil for a nil input.
func cloneMACValue(v *MACValue) *MACValue {
	if v == nil {
		return nil
	}
	return &MACValue{
		IsField: v.IsField,
		Field:   v.Field,
		IntVal:  v.IntVal,
		StrVal:  v.StrVal,
	}
}

// Reports whether two MACExpr trees are structurally equal.
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

// Reports whether two MACValue slices are element-wise equal.
func macValuesEqual(a, b []*MACValue) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !macValueEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// Reports whether two MACValue pointers are equal.
func macValueEqual(a, b *MACValue) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.IsField == b.IsField && a.Field == b.Field &&
		a.IntVal == b.IntVal && a.StrVal == b.StrVal
}
