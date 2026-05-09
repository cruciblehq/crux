package manifest

import (
	"errors"
	"testing"
)

func TestParseResourceTypeKnown(t *testing.T) {
	for _, s := range []string{"runtime", "service", "template", "widget", "affordance", "blueprint"} {
		got, err := ParseResourceType(s)
		if err != nil {
			t.Errorf("ParseResourceType(%q): %v", s, err)
		}
		if string(got) != s {
			t.Errorf("ParseResourceType(%q) = %q", s, got)
		}
	}
}

func TestParseResourceTypeInvalid(t *testing.T) {
	for _, s := range []string{"", "Runtime", "unknown", "svc"} {
		_, err := ParseResourceType(s)
		if !errors.Is(err, ErrInvalidResourceType) {
			t.Errorf("ParseResourceType(%q) err = %v, want ErrInvalidResourceType", s, err)
		}
	}
}
