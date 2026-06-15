package fcap

import (
	"errors"
	"testing"
)

func TestParseModeValid(t *testing.T) {
	for _, m := range []Mode{ModeEffective, ModeInheritable} {
		got, err := ParseMode(string(m))
		if err != nil || got != m {
			t.Fatalf("ParseMode(%q) = %q, %v", m, got, err)
		}
	}
}

func TestParseModeInvalid(t *testing.T) {
	_, err := ParseMode("bogus")
	if !errors.Is(err, ErrUnknownFcapMode) {
		t.Fatalf("err = %v, want ErrUnknownFcapMode", err)
	}
}
