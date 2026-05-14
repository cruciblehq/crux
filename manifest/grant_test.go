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

func TestDecodeGrantRefWithScalarValue(t *testing.T) {
	raw := map[string]any{
		"ns/policy": "strict",
	}
	g, err := decodeGrant(raw)
	if err != nil {
		t.Fatal(err)
	}
	if g.Source != "ns/policy" {
		t.Fatalf("source = %q, want %q", g.Source, "ns/policy")
	}
	if g.Value != "strict" {
		t.Fatalf("value = %q, want %q", g.Value, "strict")
	}
	if len(g.Args) != 0 {
		t.Fatalf("args = %v, want empty", g.Args)
	}
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeGrantRefWithNamedArgs(t *testing.T) {
	raw := map[string]any{
		"ns/policy": map[string]any{"level": "high", "mode": "strict"},
	}
	g, err := decodeGrant(raw)
	if err != nil {
		t.Fatal(err)
	}
	if g.Source != "ns/policy" {
		t.Fatalf("source = %q, want %q", g.Source, "ns/policy")
	}
	if g.Value != "" {
		t.Fatalf("value = %q, want empty", g.Value)
	}
	if g.Args["level"] != "high" {
		t.Fatalf("args[level] = %q, want %q", g.Args["level"], "high")
	}
	if g.Args["mode"] != "strict" {
		t.Fatalf("args[mode] = %q, want %q", g.Args["mode"], "strict")
	}
	if err := g.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeGrantRefWithNamedArgsIntValue(t *testing.T) {
	raw := map[string]any{
		"ns/policy": map[string]any{"level": 5},
	}
	_, err := decodeGrant(raw)
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("error = %v, want ErrInvalidAffordance", err)
	}
}

func TestGrantEncodeWithScalarValue(t *testing.T) {
	g := Grant{Source: "ns/policy", Value: "strict"}
	enc, err := g.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := enc.(map[string]any)
	if !ok {
		t.Fatalf("encode type = %T, want map[string]any", enc)
	}
	if m["ns/policy"] != "strict" {
		t.Fatalf("encoded value = %v, want %q", m["ns/policy"], "strict")
	}
}

func TestGrantEncodeWithNamedArgs(t *testing.T) {
	g := Grant{Source: "ns/policy", Args: map[string]string{"level": "high"}}
	enc, err := g.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := enc.(map[string]any)
	if !ok {
		t.Fatalf("encode type = %T, want map[string]any", enc)
	}
	inner, ok := m["ns/policy"].(map[string]any)
	if !ok {
		t.Fatalf("inner type = %T, want map[string]any", m["ns/policy"])
	}
	if inner["level"] != "high" {
		t.Fatalf("inner[level] = %v, want %q", inner["level"], "high")
	}
}

func TestGrantValidateRejectsMixed(t *testing.T) {
	g := Grant{Source: "ns/policy", Value: "strict", Args: map[string]string{"level": "high"}}
	err := g.Validate()
	if !errors.Is(err, ErrGrantArgsMixed) {
		t.Fatalf("error = %v, want ErrGrantArgsMixed", err)
	}
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want wrapped ErrInvalidGrant", err)
	}
}

func TestGrantValidateRejectsDomainWithValue(t *testing.T) {
	g := Grant{Source: ".cap effective net_admin", Value: "strict"}
	err := g.Validate()
	if !errors.Is(err, ErrDomainGrantWithArgs) {
		t.Fatalf("error = %v, want ErrDomainGrantWithArgs", err)
	}
}

func TestGrantValidateRejectsDomainWithArgs(t *testing.T) {
	g := Grant{Source: ".cap effective net_admin", Args: map[string]string{"k": "v"}}
	err := g.Validate()
	if !errors.Is(err, ErrDomainGrantWithArgs) {
		t.Fatalf("error = %v, want ErrDomainGrantWithArgs", err)
	}
}

func TestDecodeGrantRefWithUnsupportedValueType(t *testing.T) {
	raw := map[string]any{
		"ns/policy": []any{"a", "b"},
	}
	_, err := decodeGrant(raw)
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("error = %v, want ErrInvalidAffordance", err)
	}
}

func TestGrantValidateRejectsInvalidArgKey(t *testing.T) {
	g := Grant{Source: "ns/policy", Args: map[string]string{"Bad-Key": "v"}}
	err := g.Validate()
	if !errors.Is(err, ErrInvalidArgKey) {
		t.Fatalf("error = %v, want ErrInvalidArgKey", err)
	}
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("error = %v, want wrapped ErrInvalidGrant", err)
	}
}
