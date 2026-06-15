package mac

import (
	"errors"
	"testing"
)

func TestMACExprValidateAndMissingChildren(t *testing.T) {
	e := &MACExpr{Type: "and"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateAndInvalidLeft(t *testing.T) {
	e := &MACExpr{
		Type:  "and",
		Left:  &MACExpr{Type: "bogus"},
		Right: &MACExpr{Type: "cmp", Op: "=", LHS: &MACValue{IsField: true, Field: "task.uid"}, RHS: &MACValue{}},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateOrInvalidRight(t *testing.T) {
	e := &MACExpr{
		Type:  "or",
		Left:  &MACExpr{Type: "cmp", Op: "=", LHS: &MACValue{IsField: true, Field: "task.uid"}, RHS: &MACValue{}},
		Right: &MACExpr{Type: "bogus"},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateNotMissingOperand(t *testing.T) {
	e := &MACExpr{Type: "not"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateNotInvalidOperand(t *testing.T) {
	e := &MACExpr{Type: "not", Operand: &MACExpr{Type: "bogus"}}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateCmpMissing(t *testing.T) {
	e := &MACExpr{Type: "cmp"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateCmpInvalidLHS(t *testing.T) {
	e := &MACExpr{
		Type: "cmp",
		Op:   "=",
		LHS:  &MACValue{IsField: true, Field: ""},
		RHS:  &MACValue{},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateCmpInvalidRHS(t *testing.T) {
	e := &MACExpr{
		Type: "cmp",
		Op:   "=",
		LHS:  &MACValue{IsField: true, Field: "task.uid"},
		RHS:  &MACValue{IsField: true, Field: ""},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateCmpOK(t *testing.T) {
	e := &MACExpr{
		Type: "cmp",
		Op:   "=",
		LHS:  &MACValue{IsField: true, Field: "task.uid"},
		RHS:  &MACValue{IntVal: 0},
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMACExprValidateInMissing(t *testing.T) {
	e := &MACExpr{Type: "in"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateInInvalidField(t *testing.T) {
	e := &MACExpr{
		Type:   "in",
		Field:  &MACValue{IsField: true, Field: ""},
		Values: []*MACValue{{IntVal: 1}},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateInNilValue(t *testing.T) {
	e := &MACExpr{
		Type:   "in",
		Field:  &MACValue{IsField: true, Field: "task.uid"},
		Values: []*MACValue{nil},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateInInvalidValue(t *testing.T) {
	e := &MACExpr{
		Type:   "in",
		Field:  &MACValue{IsField: true, Field: "task.uid"},
		Values: []*MACValue{{IsField: true, Field: ""}},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateLikeMissing(t *testing.T) {
	e := &MACExpr{Type: "like"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateLikeInvalidField(t *testing.T) {
	e := &MACExpr{
		Type:    "like",
		Field:   &MACValue{IsField: true, Field: ""},
		Pattern: "/tmp/*",
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateLikeOK(t *testing.T) {
	e := &MACExpr{
		Type:    "like",
		Field:   &MACValue{IsField: true, Field: "file.path"},
		Pattern: "/tmp/*",
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMACExprValidateBetweenMissing(t *testing.T) {
	e := &MACExpr{Type: "between"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateBetweenInvalidField(t *testing.T) {
	e := &MACExpr{
		Type:  "between",
		Field: &MACValue{IsField: true, Field: ""},
		Low:   &MACValue{IntVal: 0},
		High:  &MACValue{IntVal: 100},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateBetweenInvalidLow(t *testing.T) {
	e := &MACExpr{
		Type:  "between",
		Field: &MACValue{IsField: true, Field: "task.uid"},
		Low:   &MACValue{IsField: true, Field: ""},
		High:  &MACValue{IntVal: 100},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateBetweenInvalidHigh(t *testing.T) {
	e := &MACExpr{
		Type:  "between",
		Field: &MACValue{IsField: true, Field: "task.uid"},
		Low:   &MACValue{IntVal: 0},
		High:  &MACValue{IsField: true, Field: ""},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateBetweenOK(t *testing.T) {
	e := &MACExpr{
		Type:  "between",
		Field: &MACValue{IsField: true, Field: "task.uid"},
		Low:   &MACValue{IntVal: 0},
		High:  &MACValue{IntVal: 100},
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMACExprValidateBittestMissing(t *testing.T) {
	e := &MACExpr{Type: "bittest"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateBittestInvalidField(t *testing.T) {
	e := &MACExpr{
		Type:  "bittest",
		Field: &MACValue{IsField: true, Field: ""},
		Mask:  &MACValue{IntVal: 7},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateBittestInvalidMask(t *testing.T) {
	e := &MACExpr{
		Type:  "bittest",
		Field: &MACValue{IsField: true, Field: "task.uid"},
		Mask:  &MACValue{IsField: true, Field: ""},
	}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}

func TestMACExprValidateBittestOK(t *testing.T) {
	e := &MACExpr{
		Type:  "bittest",
		Field: &MACValue{IsField: true, Field: "task.uid"},
		Mask:  &MACValue{IntVal: 7},
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMACExprValidateUnknownType(t *testing.T) {
	e := &MACExpr{Type: "unknown"}
	if err := e.Validate(); !errors.Is(err, ErrInvalidMACExpr) {
		t.Fatalf("err = %v, want ErrInvalidMACExpr", err)
	}
}
