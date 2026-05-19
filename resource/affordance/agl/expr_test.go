package agl

import "testing"

func fieldOp(name string) Operand {
	return Operand{IsField: true, Field: name}
}

func intOp(v uint64) Operand {
	return Operand{Value: Value{Type: ValueInt, Int: v}}
}

func TestBinaryExprString(t *testing.T) {
	e := &BinaryExpr{
		Op:    OpAnd,
		Left:  &CompareExpr{Left: fieldOp("a"), Op: CmpEq, Right: intOp(1)},
		Right: &CompareExpr{Left: fieldOp("b"), Op: CmpGt, Right: intOp(2)},
	}
	want := "(a = 1 and b > 2)"
	if got := e.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUnaryExprString(t *testing.T) {
	e := &UnaryExpr{
		Op:      OpNot,
		Operand: &CompareExpr{Left: fieldOp("a"), Op: CmpEq, Right: intOp(1)},
	}
	want := "not a = 1"
	if got := e.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCompareExprString(t *testing.T) {
	for _, op := range []CmpOp{CmpEq, CmpNeq, CmpGt, CmpGte, CmpLt, CmpLte} {
		e := &CompareExpr{Left: fieldOp("x"), Op: op, Right: intOp(0)}
		want := "x " + string(op) + " 0"
		if got := e.String(); got != want {
			t.Errorf("op %s: got %q, want %q", op, got, want)
		}
	}
}

func TestInExprString(t *testing.T) {
	e := &InExpr{
		Field:  fieldOp("p"),
		Values: []Operand{intOp(1), intOp(2), intOp(3)},
	}
	want := "p in (1, 2, 3)"
	if got := e.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLikeExprString(t *testing.T) {
	pos := &LikeExpr{Field: fieldOp("path"), Pattern: "*.so", Negated: false}
	if got, want := pos.String(), `path like "*.so"`; got != want {
		t.Errorf("positive: got %q, want %q", got, want)
	}
	neg := &LikeExpr{Field: fieldOp("path"), Pattern: "*.tmp", Negated: true}
	if got, want := neg.String(), `path not like "*.tmp"`; got != want {
		t.Errorf("negated: got %q, want %q", got, want)
	}
}

func TestBetweenExprString(t *testing.T) {
	e := &BetweenExpr{Field: fieldOp("n"), Low: intOp(1), High: intOp(10)}
	want := "n between 1 and 10"
	if got := e.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBitTestExprString(t *testing.T) {
	noVal := &BitTestExpr{Field: fieldOp("flags"), Mask: intOp(4)}
	if got, want := noVal.String(), "flags & 4"; got != want {
		t.Errorf("no value: got %q, want %q", got, want)
	}
	val := intOp(4)
	withVal := &BitTestExpr{Field: fieldOp("flags"), Mask: intOp(4), Val: &val}
	if got, want := withVal.String(), "flags & 4 = 4"; got != want {
		t.Errorf("with value: got %q, want %q", got, want)
	}
}
