package source

import (
	"strings"
	"testing"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/registry"
)

func TestCheckCache_Miss_NoEntry(t *testing.T) {
	c := openTestCache(t)

	ref := mustParseRef(t, "acme/myruntime 1.0.0")
	ver := &registry.Version{String: "1.0.0"}

	result, ok := checkCache(c, ref, ver, "sha256:doesnotexist")
	if ok {
		t.Fatal("expected cache miss, got hit")
	}
	if result != nil {
		t.Errorf("expected nil result on miss, got %v", result)
	}
}

func TestCheckCache_Miss_DigestMismatch(t *testing.T) {
	c := openTestCache(t)
	ref := mustParseRef(t, "acme/myruntime 1.0.0")
	ver := &registry.Version{String: "1.0.0"}

	// Store an entry (Put computes and stores its real digest).
	if _, err := c.Put("acme", "myruntime", "1.0.0", strings.NewReader("data")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Check with a different expected digest — should be a miss.
	result, ok := checkCache(c, ref, ver, "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if ok {
		t.Fatal("expected cache miss due to digest mismatch, got hit")
	}
	if result != nil {
		t.Errorf("expected nil result on mismatch, got %v", result)
	}
}

func TestCheckCache_Hit(t *testing.T) {
	c := openTestCache(t)
	ref := mustParseRef(t, "acme/myruntime 1.0.0")
	ver := &registry.Version{String: "1.0.0"}

	// Store an entry and capture the computed digest.
	stored, err := c.Put("acme", "myruntime", "1.0.0", strings.NewReader("archive data"))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if stored.Digest == nil {
		t.Fatal("expected non-nil digest from Put")
	}
	digest := *stored.Digest

	result, ok := checkCache(c, ref, ver, digest)
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if result == nil {
		t.Fatal("expected non-nil result on hit")
	}
	if result.Namespace != "acme" {
		t.Errorf("Namespace = %q, want %q", result.Namespace, "acme")
	}
	if result.Resource != "myruntime" {
		t.Errorf("Resource = %q, want %q", result.Resource, "myruntime")
	}
	if result.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.0.0")
	}
	if result.Digest != digest {
		t.Errorf("Digest = %q, want %q", result.Digest, digest)
	}
}

// Opens a temporary cache rooted in t.TempDir.
func openTestCache(t *testing.T) *cache.Cache {
	t.Helper()
	c, err := cache.OpenAt(t.TempDir())
	if err != nil {
		t.Fatalf("cache.OpenAt: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// Parses a reference string with a runtime context type.
func mustParseRef(t *testing.T, s string) *reference.Reference {
	t.Helper()
	ref, err := reference.Parse(s, "runtime")
	if err != nil {
		t.Fatalf("reference.Parse(%q): %v", s, err)
	}
	return ref.WithDefaults("https://hub.crucible.sh", "acme")
}
