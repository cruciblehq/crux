package mac

import "testing"

func TestCatalogNonNil(t *testing.T) {
	r := catalog()
	if r == nil {
		t.Fatal("catalog() returned nil")
	}
}

func TestCatalogHasFileOpen(t *testing.T) {
	r := catalog()
	h := r.LookupHook("file_open")
	if h == nil {
		t.Fatal("catalog() missing file_open hook")
	}
}

func TestCatalogUnknownHookNil(t *testing.T) {
	r := catalog()
	if h := r.LookupHook("definitely_not_a_hook"); h != nil {
		t.Fatalf("expected nil for unknown hook, got %+v", h)
	}
}

func TestCatalogSingleton(t *testing.T) {
	r1 := catalog()
	r2 := catalog()
	if r1 != r2 {
		t.Fatal("catalog() should return the same pointer on repeated calls")
	}
}
