package affordance

import "testing"

func TestNewSpecReturnsInitialised(t *testing.T) {
	s := NewSpec()
	if s == nil {
		t.Fatal("NewSpec() returned nil")
	}
	if s.OCI == nil {
		t.Fatal("OCI is nil")
	}
	if s.Fcap == nil {
		t.Fatal("Fcap is nil")
	}
	if s.MAC == nil {
		t.Fatal("MAC is nil")
	}
	if s.Net == nil {
		t.Fatal("Net is nil")
	}
}
