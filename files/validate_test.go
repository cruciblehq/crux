package files

import (
	"errors"
	"testing"
)

func TestValidateAbsPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"absolute clean", "/usr/bin/foo", "/usr/bin/foo", true},
		{"single segment", "/tmp", "/tmp", true},
		{"empty", "", "", false},
		{"root trailing slash", "/", "", false},
		{"nul", "/usr/\x00bin", "", false},
		{"relative", "usr/bin/foo", "", false},
		{"dot relative", "./foo", "", false},
		{"trailing slash", "/usr/bin/", "", false},
		{"not clean dotdot", "/usr/../bin/foo", "", false},
		{"not clean double slash", "/usr//bin", "", false},
		{"not clean dot", "/usr/./bin", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateAbsPath(c.in)
			if c.ok {
				if err != nil {
					t.Fatalf("ValidateAbsPath(%q) error = %v, want nil", c.in, err)
				}
				if got != c.want {
					t.Fatalf("ValidateAbsPath(%q) = %q, want %q", c.in, got, c.want)
				}
				return
			}
			if !errors.Is(err, ErrInvalidPath) {
				t.Fatalf("ValidateAbsPath(%q) error = %v, want ErrInvalidPath", c.in, err)
			}
		})
	}
}
