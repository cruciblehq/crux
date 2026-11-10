package compute

import (
	"errors"
	"testing"
)

func TestBackendForLocal(t *testing.T) {
	b, err := BackendFor(Local)
	if err != nil {
		t.Fatalf("BackendFor(Local) error = %v", err)
	}
	if b == nil {
		t.Fatal("BackendFor(Local) = nil, want a backend")
	}
	again, _ := BackendFor(Local)
	if again != b {
		t.Error("BackendFor(Local) returned a different backend on the second call")
	}
}

func TestBackendForUnknown(t *testing.T) {
	if _, err := BackendFor(Provider("bogus")); !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("BackendFor(bogus) error = %v, want ErrUnknownProvider", err)
	}
}
