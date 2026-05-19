package cap

import (
	"errors"
	"testing"
)

func TestParseModeValid(t *testing.T) {
	cases := []struct {
		in   string
		want mode
	}{
		{"full", modeFull},
		{"effective", modeEffective},
		{"inheritable", modeInheritable},
		{"permitted", modePermitted},
		{"bound", modeBound},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := parseMode(c.in)
			if err != nil {
				t.Fatalf("parseMode(%q): %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("parseMode(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestParseModeInvalid(t *testing.T) {
	cases := []string{"", "FULL", "Effective", "ambient", "permitted ", "unknown"}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			got, err := parseMode(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("parseMode(%q) err = %v, want ErrInvalidGrant", in, err)
			}
			if got != "" {
				t.Fatalf("parseMode(%q) = %q, want zero value", in, got)
			}
		})
	}
}
