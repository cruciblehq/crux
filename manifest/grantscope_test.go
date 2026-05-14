package manifest

import (
	"errors"
	"testing"
)

func TestGrantScopeValidateEmpty(t *testing.T) {
	if err := (&GrantScope{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestGrantScopeValidateInvalidPlatform(t *testing.T) {
	gs := &GrantScope{Platform: "Linux/AMD64"}
	err := gs.Validate()
	if !errors.Is(err, ErrInvalidPlatform) {
		t.Fatalf("err = %v, want ErrInvalidPlatform", err)
	}
}

func TestGrantScopeValidatePropagates(t *testing.T) {
	gs := &GrantScope{Grants: []Grant{{}}}
	err := gs.Validate()
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestGrantScopeEncodeUniversalProducesList(t *testing.T) {
	gs := &GrantScope{Grants: []Grant{
		{Source: ".cap effective net_admin"},
		{Source: "my-affordance"},
	}}
	got, err := gs.Encode()
	if err != nil {
		t.Fatal(err)
	}
	list, ok := got.([]any)
	if !ok {
		t.Fatalf("got %T, want []any", got)
	}
	if len(list) != 2 || list[0] != ".cap effective net_admin" || list[1] != "my-affordance" {
		t.Fatalf("list = %v", list)
	}
}

func TestGrantScopeEncodePlatformProducesMap(t *testing.T) {
	gs := &GrantScope{
		Platform: "linux/amd64",
		Grants:   []Grant{{Source: ".cap effective net_admin"}},
	}
	got, err := gs.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("got %T, want map", got)
	}
	if m["platform"] != "linux/amd64" {
		t.Fatalf("platform = %v", m["platform"])
	}
	entries, ok := m["grants"].([]any)
	if !ok || len(entries) != 1 {
		t.Fatalf("grants = %v", m["grants"])
	}
}

func TestGrantScopeDecodePlatformGroup(t *testing.T) {
	raw := map[string]any{
		"platform": "linux/amd64",
		"grants": []any{
			".cap effective net_admin",
			"my-affordance",
		},
	}
	var gs GrantScope
	if err := gs.Decode(raw); err != nil {
		t.Fatal(err)
	}
	if gs.Platform != "linux/amd64" {
		t.Fatalf("platform = %q", gs.Platform)
	}
	if len(gs.Grants) != 2 {
		t.Fatalf("grants len = %d", len(gs.Grants))
	}
}

func TestGrantScopeDecodeRejectsNonMap(t *testing.T) {
	err := (&GrantScope{}).Decode("string")
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("err = %v, want ErrInvalidAffordance", err)
	}
}

func TestGrantScopeDecodeRejectsMissingGrants(t *testing.T) {
	err := (&GrantScope{}).Decode(map[string]any{"platform": "linux/amd64"})
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("err = %v, want ErrInvalidAffordance", err)
	}
}
