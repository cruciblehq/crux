package manifest

import (
	"testing"

	"github.com/cruciblehq/crux/codec"
)

func TestRuntimeValidateOK(t *testing.T) {
	r := &Runtime{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeValidatePropagatesSchemaError(t *testing.T) {
	r := &Runtime{Schema: &Schema{Default: "missing"}}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRuntimeValidatePropagatesRecipeError(t *testing.T) {
	if err := (&Runtime{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRuntimeEncodeWithSchema(t *testing.T) {
	r := &Runtime{
		Schema: &Schema{Params: []Param{{Name: "version"}}},
		Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}},
	}
	raw, err := r.Encode(codec.Default())
	if err != nil {
		t.Fatal(err)
	}
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("Encode() = %T, want map[string]any", raw)
	}
	if _, ok := m["schema"]; !ok {
		t.Fatal("expected schema key in encoded runtime")
	}
}

func TestRuntimeEncodeWithoutSchema(t *testing.T) {
	r := &Runtime{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}}
	raw, err := r.Encode(codec.Default())
	if err != nil {
		t.Fatal(err)
	}
	m := raw.(map[string]any)
	if _, ok := m["schema"]; ok {
		t.Fatal("expected no schema key for nil schema")
	}
}
