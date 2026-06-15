package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

// Returns n cryptographically random bytes encoded as a lowercase hex string.
//
// Panics if the system random source fails, since that leaves no safe value to
// return and the failure is not recoverable.
func RandHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
