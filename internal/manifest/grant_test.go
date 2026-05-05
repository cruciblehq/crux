package manifest

import (
	"errors"
	"testing"
)

func TestDecodeGrantStringDomain(t *testing.T) {
	g, err := decodeGrant(".cap effective net_admin")
	if err != nil {
		t.Fatal(err)
	}
	if g.Source != ".cap effective net_admin" {
		t.Fatalf("source = %q", g.Source)
	}
	if g.IsRef() {
		t.Fatal("IsRef = true, want false")
	}
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeGrantStringReference(t *testing.T) {
	g, err := decodeGrant("my-affordance")
	if err != nil {
		t.Fatal(err)
	}
	if !g.IsRef() {
		t.Fatal("IsRef = false, want true")
	}
	if g.RefTarget() != "my-affordance" {
		t.Fatalf("RefTarget = %q", g.RefTarget())
	}
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeGrantMapDomainWithNilValue(t *testing.T) {
	raw := map[string]any{
		".cgroup io.max 8 0 rbps=1048576 wiops=5000": nil,
	}
	g, err := decodeGrant(raw)
	if err != nil {
		t.Fatal(err)
	}
	if g.Source != ".cgroup io.max 8 0 rbps=1048576 wiops=5000" {
		t.Fatalf("source = %q", g.Source)
	}
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeGrantRejectsDomainMapValue(t *testing.T) {
	raw := map[string]any{
		".cgroup io.max 8 0": map[string]any{"rbps": 1048576},
	}
	_, err := decodeGrant(raw)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("error = %v, want ErrInvalidAffordance", err)
	}
}

func TestDecodeGrantRejectsNonNilValue(t *testing.T) {
	raw := map[string]any{
		".cgroup io.max 8 0": []any{"rbps 1048576"},
	}
	_, err := decodeGrant(raw)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("error = %v, want ErrInvalidAffordance", err)
	}
}

func TestGrantValidateRejectsEmpty(t *testing.T) {
	g := Grant{}
	if err := g.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGrantValidateRejectsBadDomainSyntax(t *testing.T) {
	g := Grant{Source: ".cap"}
	err := g.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want ErrInvalidGrant", err)
	}
}

func TestGrantParseRejectsRef(t *testing.T) {
	g := Grant{Source: "my-affordance"}
	_, err := g.Parse()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want ErrInvalidGrant", err)
	}
}

func TestGrantEncodeRoundTrip(t *testing.T) {
	g := Grant{Source: ".cgroup io.max 8 0 rbps=1048576"}
	enc, err := g.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if got := enc.(string); got != g.Source {
		t.Fatalf("encode = %q, want %q", got, g.Source)
	}
}

func TestGrantParseDomain(t *testing.T) {
	g := Grant{Source: ".cap effective net_admin"}
	p, err := g.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("Parse returned nil model")
	}
	if p.Subsystem != "cap" {
		t.Fatalf("subsystem = %q, want %q", p.Subsystem, "cap")
	}
}

func TestGrantParseBadDomainSyntax(t *testing.T) {
	g := Grant{Source: ".cap"}
	_, err := g.Parse()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want ErrInvalidGrant", err)
	}
}

func TestGrantRefTargetDomainReturnsEmpty(t *testing.T) {
	g := Grant{Source: ".cap effective net_admin"}
	if got := g.RefTarget(); got != "" {
		t.Fatalf("RefTarget = %q, want \"\"", got)
	}
}

func TestGrantIsRefDomain(t *testing.T) {
	g := Grant{Source: ".cap effective net_admin"}
	if g.IsRef() {
		t.Fatal("IsRef = true, want false")
	}
}

func TestGrantDecodeString(t *testing.T) {
	var g Grant
	if err := g.Decode(".cap effective net_admin"); err != nil {
		t.Fatal(err)
	}
	if g.Source != ".cap effective net_admin" {
		t.Fatalf("source = %q", g.Source)
	}
}

func TestGrantDecodeUnsupportedType(t *testing.T) {
	err := (&Grant{}).Decode(42)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want ErrInvalidGrant", err)
	}
}

func TestGrantDecodeMapMultipleKeys(t *testing.T) {
	raw := map[string]any{
		".cap effective net_admin": nil,
		".cap effective sys_admin": nil,
	}
	err := (&Grant{}).Decode(raw)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want ErrInvalidGrant", err)
	}
}

func TestGrantDecodeMapEmpty(t *testing.T) {
	err := (&Grant{}).Decode(map[string]any{})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want ErrInvalidGrant", err)
	}
}

func TestDecodeGrantUnsupportedTypeWrapsAffordance(t *testing.T) {
	_, err := decodeGrant(42)
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("error = %v, want ErrInvalidAffordance", err)
	}
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want wrapped ErrInvalidGrant", err)
	}
}
