package manifest

import "testing"

func TestFcapModeIsValid(t *testing.T) {
	if !FcapModeEffective.IsValid() {
		t.Error("effective should be valid")
	}
	if !FcapModeInheritable.IsValid() {
		t.Error("inheritable should be valid")
	}
}

func TestFcapModeIsValidRejectsUnknown(t *testing.T) {
	if FcapMode("other").IsValid() {
		t.Error("unknown mode should be invalid")
	}
	if FcapMode("").IsValid() {
		t.Error("empty mode should be invalid")
	}
}

func TestFcapCapabilitiesGrantEffectiveAdds(t *testing.T) {
	c := &FcapCapabilities{}
	changed := c.GrantEffective([]string{"cap_net_admin", "cap_net_bind_service"})
	if !changed {
		t.Error("expected changed=true")
	}
	if len(c.Permitted) != 2 {
		t.Fatalf("Permitted len = %d, want 2", len(c.Permitted))
	}
	if !c.Effective {
		t.Error("expected Effective=true")
	}
}

func TestFcapCapabilitiesGrantEffectiveDeduplicated(t *testing.T) {
	c := &FcapCapabilities{Permitted: []string{"cap_net_admin"}, Effective: true}
	changed := c.GrantEffective([]string{"cap_net_admin"})
	if changed {
		t.Error("expected changed=false for duplicate")
	}
	if len(c.Permitted) != 1 {
		t.Fatalf("Permitted len = %d, want 1", len(c.Permitted))
	}
}

func TestFcapCapabilitiesGrantInheritableAdds(t *testing.T) {
	c := &FcapCapabilities{}
	changed := c.GrantInheritable([]string{"cap_setuid"})
	if !changed {
		t.Error("expected changed=true")
	}
	if len(c.Inheritable) != 1 {
		t.Fatalf("Inheritable len = %d, want 1", len(c.Inheritable))
	}
}

func TestFcapCapabilitiesGrantInheritableDeduplicated(t *testing.T) {
	c := &FcapCapabilities{Inheritable: []string{"cap_setuid"}}
	changed := c.GrantInheritable([]string{"cap_setuid"})
	if changed {
		t.Error("expected changed=false for duplicate")
	}
}

func TestMergeStringSliceAddsNew(t *testing.T) {
	dst := []string{"a"}
	changed := mergeStringSlice(&dst, []string{"b", "c"})
	if !changed {
		t.Error("expected changed=true")
	}
	if len(dst) != 3 {
		t.Fatalf("len = %d, want 3", len(dst))
	}
}

func TestMergeStringSliceSkipsDuplicates(t *testing.T) {
	dst := []string{"a", "b"}
	changed := mergeStringSlice(&dst, []string{"a", "b"})
	if changed {
		t.Error("expected changed=false")
	}
	if len(dst) != 2 {
		t.Fatalf("len = %d, want 2", len(dst))
	}
}

func TestMACSpecApplyAddsRule(t *testing.T) {
	s := &MACSpec{}
	r := &MACAllow{Hook: "file_open", Where: nil}
	if !s.Apply(r) {
		t.Error("expected true on first apply")
	}
	if len(s.Rules) != 1 {
		t.Fatalf("Rules len = %d, want 1", len(s.Rules))
	}
}

func TestMACSpecApplySkipsDuplicate(t *testing.T) {
	s := &MACSpec{}
	r := &MACAllow{Hook: "file_open"}
	s.Apply(r)
	if s.Apply(r) {
		t.Error("expected false for duplicate")
	}
	if len(s.Rules) != 1 {
		t.Fatalf("Rules len = %d, want 1", len(s.Rules))
	}
}

func TestMACSpecApplyClonesRule(t *testing.T) {
	s := &MACSpec{}
	r := &MACAllow{Hook: "file_open"}
	s.Apply(r)
	r.Hook = "modified"
	if s.Rules[0].Hook != "file_open" {
		t.Error("Apply should clone the rule, not store a reference")
	}
}

func TestMACExprEqualBothNil(t *testing.T) {
	if !macExprEqual(nil, nil) {
		t.Error("both nil should be equal")
	}
}

func TestMACExprEqualOneNil(t *testing.T) {
	e := &MACExpr{Type: "cmp"}
	if macExprEqual(nil, e) || macExprEqual(e, nil) {
		t.Error("one nil should not be equal")
	}
}

func TestMACExprEqualSameType(t *testing.T) {
	a := &MACExpr{Type: "and", Left: &MACExpr{Type: "cmp"}, Right: &MACExpr{Type: "cmp"}}
	b := &MACExpr{Type: "and", Left: &MACExpr{Type: "cmp"}, Right: &MACExpr{Type: "cmp"}}
	if !macExprEqual(a, b) {
		t.Error("structurally equal exprs should be equal")
	}
}

func TestMACExprEqualDifferentType(t *testing.T) {
	a := &MACExpr{Type: "and"}
	b := &MACExpr{Type: "or"}
	if macExprEqual(a, b) {
		t.Error("different type should not be equal")
	}
}

func TestMACExprEqualDifferentOp(t *testing.T) {
	a := &MACExpr{Type: "cmp", Op: "eq"}
	b := &MACExpr{Type: "cmp", Op: "ne"}
	if macExprEqual(a, b) {
		t.Error("different op should not be equal")
	}
}

func TestMACExprEqualDifferentPattern(t *testing.T) {
	a := &MACExpr{Type: "like", Pattern: "foo*"}
	b := &MACExpr{Type: "like", Pattern: "bar*"}
	if macExprEqual(a, b) {
		t.Error("different pattern should not be equal")
	}
}

func TestMACExprEqualDifferentLHS(t *testing.T) {
	lhs := func(s string) *MACValue { return &MACValue{StrVal: s} }
	a := &MACExpr{Type: "cmp", LHS: lhs("x")}
	b := &MACExpr{Type: "cmp", LHS: lhs("y")}
	if macExprEqual(a, b) {
		t.Error("different LHS should not be equal")
	}
}

func TestMACExprEqualDifferentRHS(t *testing.T) {
	rhs := func(s string) *MACValue { return &MACValue{StrVal: s} }
	a := &MACExpr{Type: "cmp", RHS: rhs("x")}
	b := &MACExpr{Type: "cmp", RHS: rhs("y")}
	if macExprEqual(a, b) {
		t.Error("different RHS should not be equal")
	}
}

func TestMACExprEqualDifferentField(t *testing.T) {
	field := func(s string) *MACValue { return &MACValue{IsField: true, Field: s} }
	a := &MACExpr{Type: "in", Field: field("uid")}
	b := &MACExpr{Type: "in", Field: field("gid")}
	if macExprEqual(a, b) {
		t.Error("different field should not be equal")
	}
}

func TestMACExprEqualDifferentMask(t *testing.T) {
	mask := func(n uint64) *MACValue { return &MACValue{IntVal: n} }
	a := &MACExpr{Type: "bittest", Mask: mask(1)}
	b := &MACExpr{Type: "bittest", Mask: mask(2)}
	if macExprEqual(a, b) {
		t.Error("different mask should not be equal")
	}
}

func TestMACExprEqualDifferentLow(t *testing.T) {
	v := func(n uint64) *MACValue { return &MACValue{IntVal: n} }
	a := &MACExpr{Type: "between", Low: v(1), High: v(10)}
	b := &MACExpr{Type: "between", Low: v(2), High: v(10)}
	if macExprEqual(a, b) {
		t.Error("different low should not be equal")
	}
}

func TestMACExprEqualDifferentHigh(t *testing.T) {
	v := func(n uint64) *MACValue { return &MACValue{IntVal: n} }
	a := &MACExpr{Type: "between", Low: v(1), High: v(10)}
	b := &MACExpr{Type: "between", Low: v(1), High: v(20)}
	if macExprEqual(a, b) {
		t.Error("different high should not be equal")
	}
}

func TestMACExprEqualDifferentLeft(t *testing.T) {
	a := &MACExpr{Type: "and", Left: &MACExpr{Type: "cmp", Op: "eq"}, Right: &MACExpr{Type: "cmp"}}
	b := &MACExpr{Type: "and", Left: &MACExpr{Type: "cmp", Op: "ne"}, Right: &MACExpr{Type: "cmp"}}
	if macExprEqual(a, b) {
		t.Error("different left child should not be equal")
	}
}

func TestMACExprEqualDifferentRight(t *testing.T) {
	a := &MACExpr{Type: "and", Left: &MACExpr{Type: "cmp"}, Right: &MACExpr{Type: "cmp", Op: "eq"}}
	b := &MACExpr{Type: "and", Left: &MACExpr{Type: "cmp"}, Right: &MACExpr{Type: "cmp", Op: "ne"}}
	if macExprEqual(a, b) {
		t.Error("different right child should not be equal")
	}
}

func TestMACExprEqualDifferentOperand(t *testing.T) {
	a := &MACExpr{Type: "not", Operand: &MACExpr{Type: "cmp", Op: "eq"}}
	b := &MACExpr{Type: "not", Operand: &MACExpr{Type: "cmp", Op: "ne"}}
	if macExprEqual(a, b) {
		t.Error("different operand should not be equal")
	}
}

func TestMACExprEqualDifferentValuesLen(t *testing.T) {
	v := &MACValue{StrVal: "x"}
	a := &MACExpr{Type: "in", Values: []*MACValue{v, v}}
	b := &MACExpr{Type: "in", Values: []*MACValue{v}}
	if macExprEqual(a, b) {
		t.Error("different values length should not be equal")
	}
}

func TestMACExprEqualDifferentValuesContent(t *testing.T) {
	a := &MACExpr{Type: "in", Values: []*MACValue{{StrVal: "x"}}}
	b := &MACExpr{Type: "in", Values: []*MACValue{{StrVal: "y"}}}
	if macExprEqual(a, b) {
		t.Error("different values content should not be equal")
	}
}

func TestMACValuesEqualBothEmpty(t *testing.T) {
	if !macValuesEqual(nil, nil) {
		t.Error("both nil should be equal")
	}
}

func TestMACValuesEqualDifferentLen(t *testing.T) {
	v := &MACValue{StrVal: "x"}
	if macValuesEqual([]*MACValue{v}, nil) {
		t.Error("different lengths should not be equal")
	}
}

func TestMACValueEqualBothNil(t *testing.T) {
	if !macValueEqual(nil, nil) {
		t.Error("both nil should be equal")
	}
}

func TestMACValueEqualOneNil(t *testing.T) {
	v := &MACValue{StrVal: "x"}
	if macValueEqual(nil, v) || macValueEqual(v, nil) {
		t.Error("one nil should not be equal")
	}
}

func TestMACValueEqualSame(t *testing.T) {
	a := &MACValue{IsField: true, Field: "uid"}
	b := &MACValue{IsField: true, Field: "uid"}
	if !macValueEqual(a, b) {
		t.Error("equal values should be equal")
	}
}

func TestMACValueEqualDifferent(t *testing.T) {
	a := &MACValue{IntVal: 1}
	b := &MACValue{IntVal: 2}
	if macValueEqual(a, b) {
		t.Error("different values should not be equal")
	}
}

func TestCloneMACExprNil(t *testing.T) {
	if cloneMACExpr(nil) != nil {
		t.Error("clone of nil should be nil")
	}
}

func TestCloneMACExpr(t *testing.T) {
	child := &MACExpr{Type: "child", Pattern: "p"}
	val := &MACValue{IsField: true, Field: "uid"}
	e := &MACExpr{
		Type:    "and",
		Left:    child,
		Right:   child,
		Operand: child,
		Op:      "eq",
		LHS:     val,
		RHS:     val,
		Field:   val,
		Values:  []*MACValue{val},
		Pattern: "pat",
		Low:     val,
		High:    val,
		Mask:    val,
	}
	got := cloneMACExpr(e)
	if got == e {
		t.Error("clone should be a distinct pointer")
	}
	if !macExprEqual(e, got) {
		t.Error("clone should be structurally equal to original")
	}
}

func TestCloneMACValueNil(t *testing.T) {
	if cloneMACValue(nil) != nil {
		t.Error("clone of nil should be nil")
	}
}

func TestCloneMACValue(t *testing.T) {
	v := &MACValue{IsField: true, Field: "gid", IntVal: 42, StrVal: "x"}
	got := cloneMACValue(v)
	if got == v {
		t.Error("clone should be a distinct pointer")
	}
	if !macValueEqual(v, got) {
		t.Error("clone should be equal to original")
	}
}
