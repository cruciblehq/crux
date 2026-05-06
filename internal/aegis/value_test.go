package aegis

import "testing"

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
		{Value{Type: ValueNone}, unknownValueType},
	}
	for _, tc := range tests {
		if got := tc.val.String(); got != tc.want {
			t.Errorf("Value{%v}.String() = %q, want %q", tc.val.Type, got, tc.want)
		}
	}
}
