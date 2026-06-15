package mac

import "github.com/cruciblehq/crux/crex"

// Typed value representing either a kernel field reference or a literal.
//
// When IsField is true, Field holds the kernel field name and the integer and
// string fields are unused. When IsField is false, either IntVal or StrVal
// carries the literal depending on the expression context.
type MACValue struct {
	IsField bool   `codec:"is_field"` // Marks this value as a field reference rather than a literal.
	Field   string `codec:"field"`    // Kernel field name.
	IntVal  uint64 `codec:"int_val"`  // Integer literal value.
	StrVal  string `codec:"str_val"`  // String literal value.
}

// Deep-clones a MACValue.
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

// Whether two MACValue pointers are equal.
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

// Whether two MACValue slices are element-wise equal.
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

// Validates the MAC value.
func (v *MACValue) Validate() error {
	if v.IsField && v.Field == "" {
		return crex.Wrapf(ErrInvalidMACValue, "field reference has empty name")
	}
	return nil
}
