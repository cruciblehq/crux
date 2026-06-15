package mac

import (
	"errors"
	"testing"
)

func cmpWhere(field string, intVal uint64) *MACExpr {
	return &MACExpr{
		Type: "cmp",
		Op:   "=",
		LHS:  &MACValue{IsField: true, Field: field},
		RHS:  &MACValue{IntVal: intVal},
	}
}

func TestMACApplyAddsRule(t *testing.T) {
	m := &Spec{}
	r := &MACAllow{Hook: "file_open"}
	if !m.Apply(r) {
		t.Fatal("expected true for new rule")
	}
	if len(m.Rules) != 1 {
		t.Fatalf("Rules len = %d", len(m.Rules))
	}
}

func TestMACApplyDeduplicates(t *testing.T) {
	m := &Spec{}
	r := &MACAllow{Hook: "file_open"}
	m.Apply(r)
	if m.Apply(r) {
		t.Fatal("expected false for duplicate rule")
	}
	if len(m.Rules) != 1 {
		t.Fatalf("Rules len = %d after duplicate", len(m.Rules))
	}
}

func TestMACApplyWithWhere(t *testing.T) {
	m := &Spec{}
	r := &MACAllow{Hook: "file_open", Where: cmpWhere("task.uid", 1000)}
	if !m.Apply(r) {
		t.Fatal("expected true for new rule with where clause")
	}
	if len(m.Rules) != 1 {
		t.Fatalf("Rules len = %d", len(m.Rules))
	}
}

func TestMACApplyWithWhereDeduplicate(t *testing.T) {
	m := &Spec{}
	r1 := &MACAllow{Hook: "file_open", Where: cmpWhere("task.uid", 1000)}
	r2 := &MACAllow{Hook: "file_open", Where: cmpWhere("task.uid", 1000)}
	m.Apply(r1)
	if m.Apply(r2) {
		t.Fatal("expected false for duplicate rule with same where clause")
	}
	if len(m.Rules) != 1 {
		t.Fatalf("Rules len = %d after duplicate", len(m.Rules))
	}
}

func TestMACApplyWithDifferentWhere(t *testing.T) {
	m := &Spec{}
	r1 := &MACAllow{Hook: "file_open", Where: cmpWhere("task.uid", 1000)}
	r2 := &MACAllow{Hook: "file_open", Where: cmpWhere("task.uid", 2000)}
	m.Apply(r1)
	if !m.Apply(r2) {
		t.Fatal("expected true for rule with different where clause")
	}
	if len(m.Rules) != 2 {
		t.Fatalf("Rules len = %d, want 2", len(m.Rules))
	}
}

func TestMACApplyWithInWhere(t *testing.T) {
	m := &Spec{}
	r := &MACAllow{
		Hook: "file_open",
		Where: &MACExpr{
			Type:   "in",
			Field:  &MACValue{IsField: true, Field: "task.uid"},
			Values: []*MACValue{{IntVal: 0}, {IntVal: 1000}},
		},
	}
	if !m.Apply(r) {
		t.Fatal("expected true for new rule")
	}
	r2 := &MACAllow{
		Hook: "file_open",
		Where: &MACExpr{
			Type:   "in",
			Field:  &MACValue{IsField: true, Field: "task.uid"},
			Values: []*MACValue{{IntVal: 0}, {IntVal: 1000}},
		},
	}
	if m.Apply(r2) {
		t.Fatal("expected false for duplicate in-rule")
	}
}

func TestMACValidateOK(t *testing.T) {
	m := &Spec{Rules: []*MACAllow{{Hook: "file_open"}}}
	if err := m.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMACValidateNilRule(t *testing.T) {
	m := &Spec{Rules: []*MACAllow{nil}}
	if err := m.Validate(); !errors.Is(err, ErrInvalidMAC) {
		t.Fatalf("err = %v, want ErrInvalidMAC", err)
	}
}
