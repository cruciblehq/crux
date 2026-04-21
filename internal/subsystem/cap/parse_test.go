package cap

import (
	"errors"
	"testing"
)

func TestParseDefaultsToFullMode(t *testing.T) {
	g, err := Parse("net_admin")
	if err != nil {
		t.Fatal(err)
	}
	if g.Mode != ModeFull {
		t.Fatalf("mode = %q, want %q", g.Mode, ModeFull)
	}
	if g.Name != "net_admin" {
		t.Fatalf("name = %q, want %q", g.Name, "net_admin")
	}
}

func TestParseAcceptsArbitraryWhitespace(t *testing.T) {
	g, err := Parse("\teffective\t net_admin\t")
	if err != nil {
		t.Fatal(err)
	}
	if g.Mode != ModeEffective {
		t.Fatalf("mode = %q, want %q", g.Mode, ModeEffective)
	}
	if g.Name != "net_admin" {
		t.Fatalf("name = %q, want %q", g.Name, "net_admin")
	}
}

func TestParseRejectsEmptyRule(t *testing.T) {
	_, err := Parse(" \t ")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsUnknownMode(t *testing.T) {
	_, err := Parse("bogus net_admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsUnknownCapability(t *testing.T) {
	_, err := Parse("effective not_a_cap")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsExtraCapabilities(t *testing.T) {
	_, err := Parse("effective net_admin sys_admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}