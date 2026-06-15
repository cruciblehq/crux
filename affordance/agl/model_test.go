package agl

import "testing"

func TestModelStringSimple(t *testing.T) {
	m := &Model{
		Subsystem: "cap",
		Args:      []Arg{{Type: ArgName, Value: "net_admin"}},
	}
	if got, want := m.String(), ".cap net_admin"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestModelStringWithKwargs(t *testing.T) {
	m := &Model{
		Subsystem: "cgroup",
		Args:      []Arg{{Type: ArgName, Value: "io.max"}, {Type: ArgInt, Value: "8"}, {Type: ArgInt, Value: "0"}},
		Kwargs: []Kwarg{
			{Key: "rbps", Value: Arg{Type: ArgInt, Value: "1048576"}},
			{Key: "wiops", Value: Arg{Type: ArgInt, Value: "5000"}},
		},
	}
	want := ".cgroup io.max 8 0 rbps=1048576 wiops=5000"
	if got := m.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestModelStringWithWhere(t *testing.T) {
	m := &Model{
		Subsystem: "seccomp",
		Args:      []Arg{{Type: ArgName, Value: "openat"}},
		Where: &CompareExpr{
			Left:  Operand{IsField: true, Field: "arg1"},
			Op:    CmpEq,
			Right: Operand{Value: Value{Type: ValueInt, Int: 0}},
		},
	}
	want := ".seccomp openat where arg1 = 0"
	if got := m.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestModelStringRoundTrip(t *testing.T) {
	srcs := []string{
		".cap net_admin",
		".cap effective net_admin",
		".cgroup io.max 8 0 rbps=1048576 wiops=5000",
		".seccomp openat where arg1 = 0",
		".rlimit nofile 1024 2048",
	}
	for _, src := range srcs {
		m1, err := Parse(src)
		if err != nil {
			t.Fatalf("Parse(%q): %v", src, err)
		}
		out := m1.String()
		m2, err := Parse(out)
		if err != nil {
			t.Fatalf("re-Parse(%q): %v", out, err)
		}
		if m2.String() != out {
			t.Errorf("round-trip diverged: %q -> %q -> %q", src, out, m2.String())
		}
	}
}
