package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestRandHex(t *testing.T) {
	for _, n := range []int{0, 1, 8, 16, 32} {
		s := RandHex(n)

		if len(s) != 2*n {
			t.Errorf("RandHex(%d) length = %d, want %d", n, len(s), 2*n)
		}

		b, err := hex.DecodeString(s)
		if err != nil {
			t.Errorf("RandHex(%d) = %q is not valid hex: %v", n, s, err)
			continue
		}
		if len(b) != n {
			t.Errorf("RandHex(%d) decoded to %d bytes, want %d", n, len(b), n)
		}
		if s != strings.ToLower(s) {
			t.Errorf("RandHex(%d) = %q, want lowercase hex", n, s)
		}
	}
}

func TestRandHexZeroIsEmpty(t *testing.T) {
	if s := RandHex(0); s != "" {
		t.Errorf("RandHex(0) = %q, want empty string", s)
	}
}
