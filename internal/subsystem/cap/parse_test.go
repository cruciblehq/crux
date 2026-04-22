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
	if len(g.Effective) != 1 || g.Effective[0] != "net_admin" {
		t.Fatalf("effective = %#v", g.Effective)
	}
	if len(g.Permitted) != 1 || g.Permitted[0] != "net_admin" {
		t.Fatalf("permitted = %#v", g.Permitted)
	}
	if len(g.Inheritable) != 1 || g.Inheritable[0] != "net_admin" {
		t.Fatalf("inheritable = %#v", g.Inheritable)
	}
	if len(g.Bounding) != 1 || g.Bounding[0] != "net_admin" {
		t.Fatalf("bounding = %#v", g.Bounding)
	}
	if len(g.Ambient) != 1 || g.Ambient[0] != "net_admin" {
		t.Fatalf("ambient = %#v", g.Ambient)
	}
}

func TestParseAcceptsArbitraryWhitespace(t *testing.T) {
	g, err := Parse("\teffective\t net_admin\t")
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Effective) != 1 || g.Effective[0] != "net_admin" {
		t.Fatalf("effective = %#v", g.Effective)
	}
	if len(g.Permitted) != 1 || g.Permitted[0] != "net_admin" {
		t.Fatalf("permitted = %#v", g.Permitted)
	}
	if len(g.Bounding) != 1 || g.Bounding[0] != "net_admin" {
		t.Fatalf("bounding = %#v", g.Bounding)
	}
	if len(g.Inheritable) != 0 {
		t.Fatalf("inheritable = %#v, want empty", g.Inheritable)
	}
	if len(g.Ambient) != 0 {
		t.Fatalf("ambient = %#v, want empty", g.Ambient)
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

func TestParseRuleDefaultsToFullMode(t *testing.T) {
	mode, name, err := ParseRule("net_admin")
	if err != nil {
		t.Fatal(err)
	}
	if mode != ModeFull {
		t.Fatalf("mode = %q, want %q", mode, ModeFull)
	}
	if name != "net_admin" {
		t.Fatalf("name = %q, want %q", name, "net_admin")
	}
}

func TestParseRuleExtractsMode(t *testing.T) {
	mode, name, err := ParseRule("effective net_admin")
	if err != nil {
		t.Fatal(err)
	}
	if mode != ModeEffective {
		t.Fatalf("mode = %q, want %q", mode, ModeEffective)
	}
	if name != "net_admin" {
		t.Fatalf("name = %q, want %q", name, "net_admin")
	}
}

func TestParseRuleRejectsEmptyRule(t *testing.T) {
	_, _, err := ParseRule(" \t ")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRuleRejectsUnknownMode(t *testing.T) {
	_, _, err := ParseRule("bogus net_admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRuleRejectsUnknownCapability(t *testing.T) {
	_, _, err := ParseRule("effective not_a_cap")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRuleRejectsExtraTokens(t *testing.T) {
	_, _, err := ParseRule("effective net_admin sys_admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

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
