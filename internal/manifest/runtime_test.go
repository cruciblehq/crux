package manifest

import "testing"

func TestRuntimeValidateOK(t *testing.T) {
	r := &Runtime{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeValidatePropagatesSchemaError(t *testing.T) {
	r := &Runtime{Schema: Schema{Default: "missing"}}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRuntimeValidatePropagatesRecipeError(t *testing.T) {
	if err := (&Runtime{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}
