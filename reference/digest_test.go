package reference

import "testing"

func mustParseDigest(t *testing.T, s string) *Digest {
	t.Helper()
	d, err := ParseDigest(s)
	if err != nil {
		t.Fatalf("unexpected error parsing %q: %v", s, err)
	}
	return d
}

// ParseDigest tests

func TestParseDigest_SHA256(t *testing.T) {
	d := mustParseDigest(t, "sha256:abc123")
	if d.Algorithm != "sha256" {
		t.Errorf("Algorithm = %q, want %q", d.Algorithm, "sha256")
	}
	if d.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", d.Hash, "abc123")
	}
}

func TestParseDigest_SHA512(t *testing.T) {
	d := mustParseDigest(t, "sha512:def456")
	if d.Algorithm != "sha512" {
		t.Errorf("Algorithm = %q, want %q", d.Algorithm, "sha512")
	}
	if d.Hash != "def456" {
		t.Errorf("Hash = %q, want %q", d.Hash, "def456")
	}
}

func TestParseDigest_Blake2b(t *testing.T) {
	d := mustParseDigest(t, "blake2b:789abc")
	if d.Algorithm != "blake2b" {
		t.Errorf("Algorithm = %q, want %q", d.Algorithm, "blake2b")
	}
	if d.Hash != "789abc" {
		t.Errorf("Hash = %q, want %q", d.Hash, "789abc")
	}
}

func TestParseDigest_NormalizesAlgorithmToLowercase(t *testing.T) {
	d := mustParseDigest(t, "SHA256:abc123")
	if d.Algorithm != "sha256" {
		t.Errorf("Algorithm = %q, want %q", d.Algorithm, "sha256")
	}
}

func TestParseDigest_NormalizesHashToLowercase(t *testing.T) {
	d := mustParseDigest(t, "sha256:ABC123")
	if d.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", d.Hash, "abc123")
	}
}

func TestParseDigest_TrimsWhitespace(t *testing.T) {
	d := mustParseDigest(t, "  sha256:abc123  ")
	if d.Algorithm != "sha256" {
		t.Errorf("Algorithm = %q, want %q", d.Algorithm, "sha256")
	}
	if d.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", d.Hash, "abc123")
	}
}

func TestParseDigest_MissingColon(t *testing.T) {
	_, err := ParseDigest("sha256abc123")
	if err == nil {
		t.Error("expected error for missing colon")
	}
}

func TestParseDigest_EmptyString(t *testing.T) {
	_, err := ParseDigest("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestParseDigest_EmptyAlgorithm(t *testing.T) {
	_, err := ParseDigest(":abc123")
	if err == nil {
		t.Error("expected error for empty algorithm")
	}
}

func TestParseDigest_EmptyHash(t *testing.T) {
	_, err := ParseDigest("sha256:")
	if err == nil {
		t.Error("expected error for empty hash")
	}
}

func TestParseDigest_OnlyColon(t *testing.T) {
	_, err := ParseDigest(":")
	if err == nil {
		t.Error("expected error for only colon")
	}
}

func TestParseDigest_MultipleColons(t *testing.T) {
	d := mustParseDigest(t, "sha256:abc:123")
	if d.Algorithm != "sha256" {
		t.Errorf("Algorithm = %q, want %q", d.Algorithm, "sha256")
	}
	if d.Hash != "abc:123" {
		t.Errorf("Hash = %q, want %q", d.Hash, "abc:123")
	}
}

func TestParseDigest_UnknownAlgorithm(t *testing.T) {
	d := mustParseDigest(t, "unknown:abc123")
	if d.Algorithm != "unknown" {
		t.Errorf("Algorithm = %q, want %q", d.Algorithm, "unknown")
	}
}

// String tests

func TestDigest_String(t *testing.T) {
	d := &Digest{Algorithm: "sha256", Hash: "abc123"}
	if d.String() != "sha256:abc123" {
		t.Errorf("String() = %q, want %q", d.String(), "sha256:abc123")
	}
}

func TestDigest_String_RoundTrip(t *testing.T) {
	original := "sha256:abc123def456"
	d := mustParseDigest(t, original)
	if d.String() != original {
		t.Errorf("String() = %q, want %q", d.String(), original)
	}
}

// Equal tests

func TestDigest_Equal_Same(t *testing.T) {
	d1 := &Digest{Algorithm: "sha256", Hash: "abc123"}
	d2 := &Digest{Algorithm: "sha256", Hash: "abc123"}
	if !d1.Equal(d2) {
		t.Error("expected digests to be equal")
	}
}

func TestDigest_Equal_DifferentAlgorithm(t *testing.T) {
	d1 := &Digest{Algorithm: "sha256", Hash: "abc123"}
	d2 := &Digest{Algorithm: "sha512", Hash: "abc123"}
	if d1.Equal(d2) {
		t.Error("expected digests to not be equal")
	}
}

func TestDigest_Equal_DifferentHash(t *testing.T) {
	d1 := &Digest{Algorithm: "sha256", Hash: "abc123"}
	d2 := &Digest{Algorithm: "sha256", Hash: "def456"}
	if d1.Equal(d2) {
		t.Error("expected digests to not be equal")
	}
}

func TestDigest_Equal_BothNil(t *testing.T) {
	var d1 *Digest
	var d2 *Digest
	if !d1.Equal(d2) {
		t.Error("expected nil digests to be equal")
	}
}

func TestDigest_Equal_LeftNil(t *testing.T) {
	var d1 *Digest
	d2 := &Digest{Algorithm: "sha256", Hash: "abc123"}
	if d1.Equal(d2) {
		t.Error("expected nil and non-nil digests to not be equal")
	}
}

func TestDigest_Equal_RightNil(t *testing.T) {
	d1 := &Digest{Algorithm: "sha256", Hash: "abc123"}
	var d2 *Digest
	if d1.Equal(d2) {
		t.Error("expected non-nil and nil digests to not be equal")
	}
}
