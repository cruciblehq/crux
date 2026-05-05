package aegis

import "testing"

func TestArgString(t *testing.T) {
	tests := []struct {
		arg  Arg
		want string
	}{
		{Arg{Type: ArgName, Value: "net_admin"}, "net_admin"},
		{Arg{Type: ArgInt, Value: "42"}, "42"},
		{Arg{Type: ArgQuantity, Value: "1Gi"}, "1Gi"},
		{Arg{Type: ArgStrASCII, Value: `hello "world"`}, `"hello \"world\""`},
		{Arg{Type: ArgStrUnicode, Value: "café"}, `u"café"`},
		{Arg{Type: ArgVar, Value: "MY_VAR"}, "$MY_VAR"},
		{Arg{Type: ArgType(99), Value: "x"}, "<unknown>"},
	}
	for _, tc := range tests {
		if got := tc.arg.String(); got != tc.want {
			t.Errorf("Arg{%v, %q}.String() = %q, want %q", tc.arg.Type, tc.arg.Value, got, tc.want)
		}
	}
}

func TestKwargString(t *testing.T) {
	k := Kwarg{Key: "rbps", Value: Arg{Type: ArgInt, Value: "1024"}}
	if got, want := k.String(), "rbps=1024"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOperandString(t *testing.T) {
	if got := (Operand{IsField: true, Field: "task.uid"}).String(); got != "task.uid" {
		t.Errorf("field operand: got %q", got)
	}
	if got := (Operand{Value: Value{Type: ValueInt, Int: 7}}).String(); got != "7" {
		t.Errorf("literal operand: got %q", got)
	}
}

func TestValueString(t *testing.T) {
	tests := []struct {
		val  Value
		want string
	}{
		{Value{Type: ValueInt, Int: 0}, "0"},
		{Value{Type: ValueInt, Int: 12345}, "12345"},
		{Value{Type: ValueStr, Str: "hi", StrEncoding: StrASCII}, `"hi"`},
		{Value{Type: ValueStr, Str: "café", StrEncoding: StrUnicode}, `u"café"`},
		{Value{Type: ValueVar, Str: "X"}, "$X"},
		{Value{Type: ValueNone}, "<none>"},
	}
	for _, tc := range tests {
		if got := tc.val.String(); got != tc.want {
			t.Errorf("Value{%v}.String() = %q, want %q", tc.val.Type, got, tc.want)
		}
	}
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

// Helpers.

func fieldOp(name string) Operand {
	return Operand{IsField: true, Field: name}
}

func intOp(v uint64) Operand {
	return Operand{Value: Value{Type: ValueInt, Int: v}}
}
