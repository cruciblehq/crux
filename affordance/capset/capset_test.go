package capset

import (
	"errors"
	"testing"
)

func TestParseKnownCap(t *testing.T) {
	got, err := Parse("net_admin")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got != NetAdmin {
		t.Fatalf("got %q, want %q", got, NetAdmin)
	}
}

func TestParseUnknownCap(t *testing.T) {
	_, err := Parse("definitely_not_a_cap")
	if !errors.Is(err, ErrUnknownCap) {
		t.Fatalf("err = %v, want ErrUnknownCap", err)
	}
}

func TestNormalize(t *testing.T) {
	got := Normalize(NetAdmin)
	if got != "CAP_NET_ADMIN" {
		t.Fatalf("Normalize(NetAdmin) = %q, want %q", got, "CAP_NET_ADMIN")
	}
}

func TestNormalizeLowercase(t *testing.T) {
	got := Normalize(Chown)
	if got != "CAP_CHOWN" {
		t.Fatalf("Normalize(Chown) = %q, want %q", got, "CAP_CHOWN")
	}
}
