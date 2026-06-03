package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

// Returns n cryptographically random bytes encoded as a lowercase hex string.
func RandHex(n int) string {
	buf := make([]byte, n)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}
