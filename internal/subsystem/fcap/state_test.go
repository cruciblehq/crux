package fcap

import (
	"errors"
	"testing"
)

func TestCloneStateReturnsNilForNilOrEmpty(t *testing.T) {
	if cloneState(nil) != nil {
		t.Fatal("cloneState(nil) != nil")
	}
	if cloneState(&State{}) != nil {
		t.Fatal("cloneState(empty) != nil")
	}
}

func TestCloneStateReturnsDeepCopy(t *testing.T) {
	state := &State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted:   []string{"net_raw"},
			Inheritable: []string{"sys_admin"},
			Effective:   true,
		},
	}}
	clone := cloneState(state)
	clone.Entries["/usr/bin/ping"].Permitted[0] = "net_admin"
	clone.Entries["/usr/bin/ping"].Inheritable[0] = "chown"
	clone.Entries["/usr/bin/ping"].Effective = false

	entry := state.Entries["/usr/bin/ping"]
	if entry.Permitted[0] != "net_raw" {
		t.Fatalf("permitted = %#v", entry.Permitted)
	}
	if entry.Inheritable[0] != "sys_admin" {
		t.Fatalf("inheritable = %#v", entry.Inheritable)
	}
	if !entry.Effective {
		t.Fatal("effective = false, want true")
	}
}

func TestStateApplyEffectiveGrant(t *testing.T) {
	state := NewState()
	changed, err := state.Apply(&Grant{Mode: ModeEffective, Path: "/usr/bin/ping", Caps: []string{"net_raw"}})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false")
	}
	entry := state.Entries["/usr/bin/ping"]
	assertEntry(t, entry, []string{"net_raw"}, nil, true)

	changed, err = state.Apply(&Grant{Mode: ModeEffective, Path: "/usr/bin/ping", Caps: []string{"net_raw"}})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("changed = true on duplicate apply")
	}
}

func TestStateApplyInheritableGrant(t *testing.T) {
	state := NewState()
	changed, err := state.Apply(&Grant{Mode: ModeInheritable, Path: "/usr/bin/ping", Caps: []string{"net_raw"}})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false")
	}
	assertEntry(t, state.Entries["/usr/bin/ping"], nil, []string{"net_raw"}, false)
}

func TestStateApplyInitializesNilEntriesMap(t *testing.T) {
	state := &State{}
	changed, err := state.Apply(&Grant{Mode: ModeEffective, Path: "/usr/bin/ping", Caps: []string{"net_raw"}})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("changed = false")
	}
	assertEntry(t, state.Entries["/usr/bin/ping"], []string{"net_raw"}, nil, true)
}

func TestStateApplyUnknownModeReturnsError(t *testing.T) {
	state := NewState()
	changed, err := state.Apply(&Grant{Mode: Mode("bogus"), Path: "/usr/bin/ping", Caps: []string{"net_raw"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
	if changed {
		t.Fatal("changed = true")
	}
	if cloneState(state) != nil {
		t.Fatal("state changed on invalid mode")
	}
}

func TestStateMergeDeduplicatesAndOrsEffective(t *testing.T) {
	state := NewState()
	state.Merge(&State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted:   []string{"net_raw"},
			Inheritable: []string{"sys_admin"},
			Effective:   false,
		},
	}})
	state.Merge(&State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted:   []string{"net_raw"},
			Inheritable: []string{"sys_admin", "chown"},
			Effective:   true,
		},
	}})

	assertEntry(t, state.Entries["/usr/bin/ping"], []string{"net_raw"}, []string{"sys_admin", "chown"}, true)
}

func TestStateMergeInitializesNilEntriesMap(t *testing.T) {
	state := &State{}
	state.Merge(&State{Entries: map[string]*Capabilities{
		"/usr/bin/ping": {
			Permitted:   []string{"net_raw"},
			Inheritable: []string{"sys_admin"},
			Effective:   true,
		},
	}})

	assertEntry(t, state.Entries["/usr/bin/ping"], []string{"net_raw"}, []string{"sys_admin"}, true)
}

func assertEntry(t *testing.T, entry *Capabilities, permitted, inheritable []string, effective bool) {
	t.Helper()
	if entry == nil {
		t.Fatal("entry = nil")
	}
	assertCaps(t, "permitted", entry.Permitted, permitted)
	assertCaps(t, "inheritable", entry.Inheritable, inheritable)
	if entry.Effective != effective {
		t.Fatalf("effective = %v, want %v", entry.Effective, effective)
	}
}

func assertCaps(t *testing.T, name string, got, want []string) {
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
