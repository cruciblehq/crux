package cap

import "testing"

func TestCloneStateReturnsNilForNilOrEmpty(t *testing.T) {
	if cloneState(nil) != nil {
		t.Fatal("cloneState(nil) != nil")
	}
	if cloneState(&State{}) != nil {
		t.Fatal("cloneState(empty) != nil")
	}
}

func TestCloneStateReturnsDeepCopy(t *testing.T) {
	state := &State{
		Effective:   []string{"net_admin"},
		Permitted:   []string{"net_admin"},
		Inheritable: []string{"net_admin"},
		Bounding:    []string{"net_admin"},
		Ambient:     []string{"net_admin"},
	}
	clone := cloneState(state)
	clone.Effective[0] = "sys_admin"
	clone.Permitted[0] = "sys_admin"
	clone.Inheritable[0] = "sys_admin"
	clone.Bounding[0] = "sys_admin"
	clone.Ambient[0] = "sys_admin"

	if state.Effective[0] != "net_admin" {
		t.Fatalf("effective = %#v", state.Effective)
	}
	if state.Permitted[0] != "net_admin" {
		t.Fatalf("permitted = %#v", state.Permitted)
	}
	if state.Inheritable[0] != "net_admin" {
		t.Fatalf("inheritable = %#v", state.Inheritable)
	}
	if state.Bounding[0] != "net_admin" {
		t.Fatalf("bounding = %#v", state.Bounding)
	}
	if state.Ambient[0] != "net_admin" {
		t.Fatalf("ambient = %#v", state.Ambient)
	}
}

func TestStateApplyByMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        Mode
		effective   []string
		permitted   []string
		inheritable []string
		bounding    []string
		ambient     []string
	}{
		{name: "full", mode: ModeFull, effective: []string{"net_admin"}, permitted: []string{"net_admin"}, inheritable: []string{"net_admin"}, bounding: []string{"net_admin"}, ambient: []string{"net_admin"}},
		{name: "effective", mode: ModeEffective, effective: []string{"net_admin"}, permitted: []string{"net_admin"}, bounding: []string{"net_admin"}},
		{name: "inheritable", mode: ModeInheritable, permitted: []string{"net_admin"}, inheritable: []string{"net_admin"}, bounding: []string{"net_admin"}, ambient: []string{"net_admin"}},
		{name: "permitted", mode: ModePermitted, permitted: []string{"net_admin"}, bounding: []string{"net_admin"}},
		{name: "bound", mode: ModeBound, bounding: []string{"net_admin"}},
	}

	for _, test := range tests {
		state := NewState()
		changed, err := state.Apply(&Grant{Mode: test.mode, Name: "net_admin"})
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if !changed {
			t.Fatalf("%s: changed = false", test.name)
		}
		assertSlice(t, test.name+" effective", state.Effective, test.effective)
		assertSlice(t, test.name+" permitted", state.Permitted, test.permitted)
		assertSlice(t, test.name+" inheritable", state.Inheritable, test.inheritable)
		assertSlice(t, test.name+" bounding", state.Bounding, test.bounding)
		assertSlice(t, test.name+" ambient", state.Ambient, test.ambient)

		changed, err = state.Apply(&Grant{Mode: test.mode, Name: "net_admin"})
		if err != nil {
			t.Fatalf("%s second apply: %v", test.name, err)
		}
		if changed {
			t.Fatalf("%s second apply changed = true", test.name)
		}
	}
}

func TestStateApplyUnknownModeFallsBackToFull(t *testing.T) {
	state := NewState()
	changed, err := state.Apply(&Grant{Mode: Mode("bogus"), Name: "net_admin"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false")
	}
	assertSlice(t, "effective", state.Effective, []string{"net_admin"})
	assertSlice(t, "permitted", state.Permitted, []string{"net_admin"})
	assertSlice(t, "inheritable", state.Inheritable, []string{"net_admin"})
	assertSlice(t, "bounding", state.Bounding, []string{"net_admin"})
	assertSlice(t, "ambient", state.Ambient, []string{"net_admin"})
}

func TestStateMergeDeduplicatesBySet(t *testing.T) {
	state := NewState()
	state.Merge(&State{
		Effective:   []string{"net_admin"},
		Permitted:   []string{"net_admin"},
		Inheritable: []string{"net_admin"},
		Bounding:    []string{"net_admin"},
		Ambient:     []string{"net_admin"},
	})
	state.Merge(&State{
		Effective:   []string{"net_admin", "sys_admin"},
		Permitted:   []string{"net_admin", "sys_admin"},
		Inheritable: []string{"net_admin", "sys_admin"},
		Bounding:    []string{"net_admin", "sys_admin"},
		Ambient:     []string{"net_admin", "sys_admin"},
	})

	assertSlice(t, "effective", state.Effective, []string{"net_admin", "sys_admin"})
	assertSlice(t, "permitted", state.Permitted, []string{"net_admin", "sys_admin"})
	assertSlice(t, "inheritable", state.Inheritable, []string{"net_admin", "sys_admin"})
	assertSlice(t, "bounding", state.Bounding, []string{"net_admin", "sys_admin"})
	assertSlice(t, "ambient", state.Ambient, []string{"net_admin", "sys_admin"})
}

func assertSlice(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s len = %d, want %d: %#v", name, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %q, want %q", name, i, got[i], want[i])
		}
	}
}