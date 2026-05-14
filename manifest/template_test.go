package manifest

import "testing"

func TestTemplateValidateOK(t *testing.T) {
	if err := (&Template{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateValidatePropagatesSchemaError(t *testing.T) {
	tpl := &Template{Schema: &Schema{Default: "missing"}}
	if err := tpl.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
