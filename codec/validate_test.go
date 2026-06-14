package codec

import "testing"

// Test fixture that validates the Name field is non-empty.
type validatable struct {
	Name string `codec:"name"` // Must be non-empty to pass validation.
}

// Returns an error if Name is empty.
func (v *validatable) Validate() error {
	if v.Name == "" {
		return errNameRequired
	}
	return nil
}

var errNameRequired = errorString("name is required")

// Error type backed by a string literal.
type errorString string

// Returns the error message string.
func (e errorString) Error() string { return string(e) }

func TestValidate(t *testing.T) {
	v := &validatable{Name: "ok"}
	if err := Validate(v); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_Error(t *testing.T) {
	v := &validatable{}
	if err := Validate(v); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidate_NotValidatable(t *testing.T) {
	s := &sample{Name: "x"}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for non-Validatable type")
	}
}
