package manifest

import (
	"errors"
	"testing"
)

func TestParseProviderTypeAWS(t *testing.T) {
	pt, err := ParseProviderType("aws")
	if err != nil {
		t.Fatal(err)
	}
	if pt != ProviderTypeAWS {
		t.Fatalf("pt = %q, want %q", pt, ProviderTypeAWS)
	}
}

func TestParseProviderTypeLocal(t *testing.T) {
	pt, err := ParseProviderType("local")
	if err != nil {
		t.Fatal(err)
	}
	if pt != ProviderTypeLocal {
		t.Fatalf("pt = %q, want %q", pt, ProviderTypeLocal)
	}
}

func TestParseProviderTypeUnknown(t *testing.T) {
	_, err := ParseProviderType("gcp")
	if !errors.Is(err, ErrInvalidProviderType) {
		t.Fatalf("err = %v, want ErrInvalidProviderType", err)
	}
}

func TestParseProviderTypeEmpty(t *testing.T) {
	_, err := ParseProviderType("")
	if !errors.Is(err, ErrInvalidProviderType) {
		t.Fatalf("err = %v, want ErrInvalidProviderType", err)
	}
}
