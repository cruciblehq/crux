package manifest

import (
	"errors"
	"testing"
)

func TestRecipeValidateOK(t *testing.T) {
	r := &Recipe{Stages: []Stage{
		{Name: "a", Steps: []Step{{Run: "x"}}},
		{Steps: []Step{{Run: "y"}}},
	}}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRecipeValidateNoStages(t *testing.T) {
	err := (&Recipe{}).Validate()
	if !errors.Is(err, ErrMissingOutputStage) {
		t.Fatalf("err = %v, want ErrMissingOutputStage", err)
	}
}

func TestRecipeValidateDuplicateStageName(t *testing.T) {
	r := &Recipe{Stages: []Stage{
		{Name: "a", Steps: []Step{{Run: "x"}}},
		{Name: "a", Steps: []Step{{Run: "y"}}},
	}}
	err := r.Validate()
	if !errors.Is(err, ErrDuplicateStageName) {
		t.Fatalf("err = %v, want ErrDuplicateStageName", err)
	}
}

func TestRecipeValidatePropagatesStageError(t *testing.T) {
	r := &Recipe{Stages: []Stage{{Name: "1"}}}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRecipeOutputStage(t *testing.T) {
	first := Stage{Name: "a", Steps: []Step{{Run: "x"}}}
	last := Stage{Name: "b", Steps: []Step{{Run: "y"}}}
	r := &Recipe{Stages: []Stage{first, last}}
	got := r.OutputStage()
	if got == nil || got.Name != "b" {
		t.Fatalf("OutputStage = %+v, want stage b", got)
	}
}

func TestRecipeOutputStageEmpty(t *testing.T) {
	if got := (&Recipe{}).OutputStage(); got != nil {
		t.Fatalf("OutputStage = %+v, want nil", got)
	}
}
