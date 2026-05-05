package manifest

import (
	"errors"
	"testing"
)

func TestWidgetValidateOK(t *testing.T) {
	w := &Widget{Main: "index.js"}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestWidgetValidateMissingMain(t *testing.T) {
	err := (&Widget{}).Validate()
	if !errors.Is(err, ErrMissingMain) {
		t.Fatalf("err = %v, want ErrMissingMain", err)
	}
}

func TestWidgetValidatePropagatesSchemaError(t *testing.T) {
	w := &Widget{Main: "index.js", Schema: Schema{Default: "missing"}}
	if err := w.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
