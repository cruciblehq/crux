package reference

import (
	"strings"
)

// Content-addressable digest for resource verification.
//
// Digests ensure immutability and integrity of referenced resources. When a
// digest is present, the reference is considered "frozen" and always refers
// to the exact same content.
//
// This type only validates the format (algorithm:hash). Actual hash validation
// against file contents is performed at a higher layer.
type Digest struct {
	Algorithm string // Cryptographic hash algorithm (e.g., "sha256").
	Hash      string // Hex-encoded hash value.
}

// Parses a digest string in the format "algorithm:hash".
//
// Only validates the format, not the algorithm or hash length. Algorithm and
// hash are normalized to lowercase.
func ParseDigest(s string) (*Digest, error) {
	s = strings.TrimSpace(s)

	colonIdx := strings.Index(s, ":")
	if colonIdx == -1 {
		return nil, wrap(ErrInvalidDigest, ErrMissingDigestColon)
	}

	algorithm := strings.ToLower(s[:colonIdx])
	hash := strings.ToLower(s[colonIdx+1:])

	if algorithm == "" {
		return nil, wrap(ErrInvalidDigest, ErrEmptyDigestAlgorithm)
	}

	if hash == "" {
		return nil, wrap(ErrInvalidDigest, ErrEmptyDigestHash)
	}

	return &Digest{
		Algorithm: algorithm,
		Hash:      hash,
	}, nil
}

// Returns the canonical string representation (algorithm:hash).
func (d *Digest) String() string {
	return d.Algorithm + ":" + d.Hash
}

// Whether two digests are identical.
func (d *Digest) Equal(other *Digest) bool {
	if d == nil || other == nil {
		return d == other
	}
	return d.Algorithm == other.Algorithm && d.Hash == other.Hash
}
