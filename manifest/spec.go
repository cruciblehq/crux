package manifest

import (
	"slices"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Compiled runtime model for a container.
//
// Produced by the affordance builder during blueprint build. Encodes the full
// enforcement state derived from all affordance grants: OCI runtime constraints,
// file capabilities, MAC hook rules, and resource allocations. No grant
// references remain; the spec is the terminal, static artifact placed in the
// plan for use at deploy time.
type Spec struct {
	OCI       *specs.Spec    `codec:"oci"`                 // OCI runtime spec (capabilities, seccomp, namespaces, cgroup resources, etc.).
	Fcap      *FcapSpec      `codec:"fcap,omitempty"`      // File capabilities per binary path.
	MAC       *MACSpec       `codec:"mac,omitempty"`       // MAC (LSM) hook allow rules.
	Provision *ProvisionSpec `codec:"provision,omitempty"` // Resource allocations the workload requires.
}

// Accumulated file capability spec, keyed by binary path.
type FcapSpec struct {
	Entries map[string]*FcapCapabilities `codec:"entries"` // Per-binary file capabilities.
}

// File capabilities for a single binary path.
//
// The kernel uses these fields during execve to compute the new process's
// effective capability set.
type FcapCapabilities struct {
	Permitted   []string `codec:"permitted"`   // Capabilities to grant as file-permitted.
	Inheritable []string `codec:"inheritable"` // Capabilities to grant as file-inheritable.
	Effective   bool     `codec:"effective"`   // If true, all granted file-permitted capabilities become immediately effective.
}

// Selects how file capabilities are granted on a binary.
type FcapMode string

const (
	FcapModeEffective   FcapMode = "effective"   // File-permitted + effective bit. Caps are immediately effective after exec.
	FcapModeInheritable FcapMode = "inheritable" // File-inheritable. Caps only effective if caller holds them in inheritable set.
)

// Whether m is a recognised mode value.
func (m FcapMode) IsValid() bool {
	return m == FcapModeEffective || m == FcapModeInheritable
}

// Accumulated MAC grant spec.
type MACSpec struct {
	Rules []*MACAllow `codec:"rules"` // Granted LSM hook allow rules.
}

// Subsystem-specific rule expression for MAC grants.
type MACAllow struct {
	Hook  string   `codec:"hook"`  // Kernel LSM hook name.
	Where *MACExpr `codec:"where"` // Where-clause expression tree. Nil when the grant is unconditional.
}

// Normalised expression node in the MAC subsystem model.
//
// Type selects the expression variant and determines which fields are
// populated. For binary nodes ("and", "or") Left and Right are set; for
// unary ("not") only Operand; for comparisons ("cmp") Op, LHS, and RHS;
// and so on for "in", "like", "between", and "bittest".
type MACExpr struct {
	Type    string      `codec:"type"`    // Expression type discriminator.
	Left    *MACExpr    `codec:"left"`    // Left child for binary expressions.
	Right   *MACExpr    `codec:"right"`   // Right child for binary expressions.
	Operand *MACExpr    `codec:"operand"` // Operand for unary expressions.
	Op      string      `codec:"op"`      // Operator for comparison expressions.
	LHS     *MACValue   `codec:"lhs"`     // Left-hand side value for comparison expressions.
	RHS     *MACValue   `codec:"rhs"`     // Right-hand side value for comparison expressions.
	Field   *MACValue   `codec:"field"`   // Field value for field-based expressions.
	Values  []*MACValue `codec:"values"`  // Slice of values for multi-value expressions.
	Pattern string      `codec:"pattern"` // Pattern for pattern-matching expressions.
	Low     *MACValue   `codec:"low"`     // Low value for between expressions.
	High    *MACValue   `codec:"high"`    // High value for between expressions.
	Mask    *MACValue   `codec:"mask"`    // Mask value for bit-test expressions.
}

// Typed value representing a field reference or literal.
type MACValue struct {
	IsField bool   `codec:"is_field"` // Whether this value is a field reference (true) or a literal (false).
	Field   string `codec:"field"`    // Field name when IsField is true.
	IntVal  uint64 `codec:"int_val"`  // Integer literal value when IsField is false.
	StrVal  string `codec:"str_val"`  // String literal value when IsField is false.
}

// Allocations the workload requires from the platform.
type ProvisionSpec struct {
	CPU    uint64 `codec:"cpu,omitempty"`    // Allocated CPU capacity in millicores.
	Memory uint64 `codec:"memory,omitempty"` // Allocated memory in bytes.
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
