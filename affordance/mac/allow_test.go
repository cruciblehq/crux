package mac

import (
	"errors"
	"testing"
)

func TestMACAllowValidateEmptyHook(t *testing.T) {
	a := &MACAllow{Hook: ""}
	if err := a.Validate(); !errors.Is(err, ErrInvalidMACAllow) {
		t.Fatalf("err = %v, want ErrInvalidMACAllow", err)
	}
}

func TestMACAllowValidateNilWhere(t *testing.T) {
	a := &MACAllow{Hook: "file_open"}
	if err := a.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMACAllowValidateInvalidWhere(t *testing.T) {
	a := &MACAllow{
		Hook:  "file_open",
		Where: &MACExpr{Type: "bogus"},
	}
	if err := a.Validate(); !errors.Is(err, ErrInvalidMACAllow) {
		t.Fatalf("err = %v, want ErrInvalidMACAllow", err)
	}
}
