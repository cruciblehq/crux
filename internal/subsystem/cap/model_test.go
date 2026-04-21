package cap

import (
	"errors"
	"testing"
)

func TestParseModeAcceptsKnownModes(t *testing.T) {
	tests := []Mode{
		ModeFull,
		ModeEffective,
		ModeInheritable,
		ModePermitted,
		ModeBound,
	}

	for _, mode := range tests {
		got, err := ParseMode(string(mode))
		if err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
		if got != mode {
			t.Fatalf("mode %q: got %q", mode, got)
		}
	}
}

func TestParseModeRejectsUnknownMode(t *testing.T) {
	_, err := ParseMode("bogus")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}