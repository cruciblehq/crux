package manifest

import (
	"errors"
	"testing"
)

func TestAffordanceValidateOK(t *testing.T) {
	a := &Affordance{
		Scopes: []GrantScope{
			{Grants: []Grant{{Source: ".cap effective net_admin"}}},
			{Platform: "linux/amd64", Grants: []Grant{{Source: "my-affordance"}}},
		},
	}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAffordanceValidatePropagatesGrantError(t *testing.T) {
	a := &Affordance{Scopes: []GrantScope{{Grants: []Grant{{}}}}}
	err := a.Validate()
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("err = %v, want ErrInvalidAffordance", err)
	}
}

func TestAffordanceValidatePropagatesSchemaError(t *testing.T) {
	a := &Affordance{Schema: Schema{Default: "missing"}}
	err := a.Validate()
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("err = %v, want ErrInvalidAffordance", err)
	}
}

func TestAffordanceEncodeFlattensUniversal(t *testing.T) {
	a := &Affordance{Scopes: []GrantScope{
		{Grants: []Grant{{Source: ".cap effective net_admin"}, {Source: "ref-affordance"}}},
	}}
	got, err := a.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m := got.(map[string]any)
	list, ok := m["grants"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("grants = %v", m["grants"])
	}
	if list[0] != ".cap effective net_admin" || list[1] != "ref-affordance" {
		t.Fatalf("list = %v", list)
	}
}

func TestAffordanceEncodePlatformAsGroup(t *testing.T) {
	a := &Affordance{Scopes: []GrantScope{
		{Platform: "linux/amd64", Grants: []Grant{{Source: ".cap effective net_admin"}}},
	}}
	got, err := a.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m := got.(map[string]any)
	list := m["grants"].([]any)
	if len(list) != 1 {
		t.Fatalf("len = %d", len(list))
	}
	entry := list[0].(map[string]any)
	if entry["platform"] != "linux/amd64" {
		t.Fatalf("platform = %v", entry["platform"])
	}
}

func TestAffordanceDecodeUniversalAndPlatform(t *testing.T) {
	raw := map[string]any{
		"grants": []any{
			".cap effective net_admin",
			"my-affordance",
			map[string]any{
				"platform": "linux/amd64",
				"grants":   []any{".cap effective sys_admin"},
			},
		},
	}
	var a Affordance
	if err := a.Decode(raw); err != nil {
		t.Fatal(err)
	}
	if len(a.Scopes) != 2 {
		t.Fatalf("scopes = %d", len(a.Scopes))
	}
	if a.Scopes[0].Platform != "" || len(a.Scopes[0].Grants) != 2 {
		t.Fatalf("universal scope wrong: %+v", a.Scopes[0])
	}
	if a.Scopes[1].Platform != "linux/amd64" || len(a.Scopes[1].Grants) != 1 {
		t.Fatalf("platform scope wrong: %+v", a.Scopes[1])
	}
}

func TestAffordanceDecodeRejectsNonMap(t *testing.T) {
	err := (&Affordance{}).Decode("string")
	if !errors.Is(err, ErrInvalidAffordance) {
		t.Fatalf("err = %v, want ErrInvalidAffordance", err)
	}
}

func TestAffordanceDecodeEmptyGrants(t *testing.T) {
	var a Affordance
	if err := a.Decode(map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if len(a.Scopes) != 0 {
		t.Fatalf("expected no scopes, got %d", len(a.Scopes))
	}
}
