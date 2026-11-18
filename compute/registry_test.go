package compute

import (
	"errors"
	"testing"
)

func TestProviderForLocal(t *testing.T) {
	p, err := ProviderFor(Local)
	if err != nil {
		t.Fatalf("ProviderFor(Local) error = %v, want nil", err)
	}
	if _, ok := p.(localProvider); !ok {
		t.Fatalf("ProviderFor(Local) = %T, want localProvider", p)
	}
}

func TestProviderForUnknown(t *testing.T) {
	if _, err := ProviderFor(ProviderName("bogus")); !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("ProviderFor(bogus) error = %v, want ErrUnknownProvider", err)
	}
}
