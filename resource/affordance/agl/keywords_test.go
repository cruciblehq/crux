package agl

import "testing"

func TestIsKeyword(t *testing.T) {
	for _, s := range []string{"where", "and", "or", "not", "in", "like", "between"} {
		if !isKeyword(s) {
			t.Errorf("isKeyword(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"WHERE", "And", "name", "foo", ""} {
		if isKeyword(s) {
			t.Errorf("isKeyword(%q) = true, want false", s)
		}
	}
}
