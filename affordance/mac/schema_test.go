package mac

import "testing"

func TestNewRegistryIsEmpty(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if h := r.LookupHook("file_open"); h != nil {
		t.Fatalf("expected nil for unknown hook, got %+v", h)
	}
}

func TestAddHookAndLookup(t *testing.T) {
	r := NewRegistry()
	h := Hook{Name: "test_hook", Fields: map[string]Field{
		"f": {Name: "f", Type: TypeString},
	}}
	r.AddHook(h)
	got := r.LookupHook("test_hook")
	if got == nil {
		t.Fatal("LookupHook returned nil after AddHook")
	}
	if got.Name != "test_hook" {
		t.Fatalf("Name = %q, want %q", got.Name, "test_hook")
	}
}

func TestLookupUnknownReturnsNil(t *testing.T) {
	r := NewRegistry()
	if h := r.LookupHook("no_such_hook"); h != nil {
		t.Fatalf("expected nil, got %+v", h)
	}
}

func TestAddHookReplaces(t *testing.T) {
	r := NewRegistry()
	r.AddHook(Hook{Name: "hook_a", Sleepable: false})
	r.AddHook(Hook{Name: "hook_a", Sleepable: true})
	got := r.LookupHook("hook_a")
	if got == nil {
		t.Fatal("hook missing after replacement")
	}
	if !got.Sleepable {
		t.Fatal("second AddHook should have replaced the first")
	}
}
