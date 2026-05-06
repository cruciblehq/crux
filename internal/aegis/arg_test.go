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
		{Arg{Type: ArgType(99), Value: "x"}, unknownArgType},
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
