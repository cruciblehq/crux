package manifest

import "slices"

// Accumulated file capability spec, keyed by binary path.
//
// Multiple affordances targeting the same binary path are merged into a single
// entry so the resulting map holds the minimal set of grants needed by the
// workload.
type FcapSpec struct {

	// Per-binary file capabilities.
	//
	// Keyed by absolute path in the container filesystem.
	Entries map[string]*FcapCapabilities `codec:"entries"`
}

// File capabilities for a single binary path.
//
// The kernel uses these fields during execve to compute the new process's
// effective capability set.
type FcapCapabilities struct {

	// Capabilities to grant as file-permitted.
	Permitted []string `codec:"permitted"`

	// Capabilities to grant as file-inheritable.
	Inheritable []string `codec:"inheritable"`

	// Immediately activates the file-permitted capabilities after execve.
	//
	// When true, the kernel sets the file effective bit so the new process does
	// not need to raise the capabilities itself.
	Effective bool `codec:"effective"`
}

// Selects how file capabilities are granted on a binary.
//
// Effective mode also sets the file effective bit, which causes the granted
// capabilities to be active immediately after execve without the new process
// having to raise them itself. Inheritable mode requires the calling process
// to already hold the capabilities in its inheritable set.
type FcapMode string

const (

	// File-permitted plus effective bit.
	//
	// Capabilities are immediately active after execve without the process
	// needing to raise them.
	FcapModeEffective FcapMode = "effective"

	// File-inheritable.
	//
	// Capabilities take effect only if the caller already holds them in its
	// inheritable set.
	FcapModeInheritable FcapMode = "inheritable"
)

// Whether m is a recognised mode value.
func (m FcapMode) IsValid() bool {
	return m == FcapModeEffective || m == FcapModeInheritable
}

// Accumulated MAC grant spec.
//
// Built up by successive calls to Apply during affordance evaluation.
// Duplicate rules are suppressed so the slice contains only the minimal set
// of distinct hook allows required by the workload.
type MACSpec struct {
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

// Grants file-permitted capabilities and sets the effective bit.
//
// After execve the capabilities are immediately effective in the new process.
// Returns true if any state was changed.
func (c *FcapCapabilities) GrantEffective(caps []string) bool {
	changed := mergeStringSlice(&c.Permitted, caps)
	if !c.Effective {
		c.Effective = true
		changed = true
	}
	return changed
}

// Grants file-inheritable capabilities.
//
// The capabilities only take effect if the calling process also holds
// them in its inheritable set. Returns true if any state was changed.
func (c *FcapCapabilities) GrantInheritable(caps []string) bool {
	return mergeStringSlice(&c.Inheritable, caps)
}

// Appends elements from src that are not already in dst.
//
// Returns true if any element was added.
func mergeStringSlice(dst *[]string, src []string) bool {
	changed := false
	for _, s := range src {
		found := slices.Contains(*dst, s)
		if !found {
			*dst = append(*dst, s)
			changed = true
		}
	}
	return changed
}

// Appends allow to the spec's rule list if it is not already present.
//
// Returns true if the rule was added (false means a rule with the same hook
// and where clause already existed, and the call was a no-op).
func (s *MACSpec) Apply(r *MACAllow) bool {
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
