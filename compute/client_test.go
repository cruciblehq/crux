package compute

import (
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestPathUnder(t *testing.T) {
	tests := []struct {
		name, path, dir string
		want            bool
	}{
		{"exact match", "/a/b", "/a/b", true},
		{"child", "/a/b/c", "/a/b", true},
		{"deep child", "/a/b/c/d", "/a/b", true},
		{"sibling", "/a/c", "/a/b", false},
		{"parent", "/a", "/a/b", false},
		{"prefix but not path", "/a/bc", "/a/b", false},
	}

	for _, tc := range tests {
		got := pathUnder(tc.path, tc.dir)
		if got != tc.want {
			t.Errorf("%s: pathUnder(%q, %q) = %v; want %v", tc.name, tc.path, tc.dir, got, tc.want)
		}
	}
}

func TestApplyWhiteout_Regular(t *testing.T) {
	entries := map[string]ocispec.Descriptor{
		"/a/b": {Digest: "sha256:aabb"},
		"/a/c": {Digest: "sha256:ccdd"},
	}
	applyWhiteout(entries, "/a/.wh.b", ".wh.b")
	if _, ok := entries["/a/b"]; ok {
		t.Error("expected /a/b to be deleted")
	}
	if _, ok := entries["/a/c"]; !ok {
		t.Error("expected /a/c to remain")
	}
}

func TestApplyWhiteout_Opaque(t *testing.T) {
	entries := map[string]ocispec.Descriptor{
		"/a/b":   {Digest: "sha256:aabb"},
		"/a/b/c": {Digest: "sha256:bbcc"},
		"/x/y":   {Digest: "sha256:xxyy"},
	}
	applyWhiteout(entries, "/a/b/.wh..wh..opq", ".wh..wh..opq")
	if _, ok := entries["/a/b"]; ok {
		t.Error("opaque whiteout should delete the dir itself")
	}
	if _, ok := entries["/a/b/c"]; ok {
		t.Error("opaque whiteout should delete children")
	}
	if _, ok := entries["/x/y"]; !ok {
		t.Error("opaque whiteout should not delete unrelated entries")
	}
}

func TestApplyWhiteout_MissingTarget(t *testing.T) {
	entries := map[string]ocispec.Descriptor{
		"/a/b": {Digest: "sha256:aabb"},
	}
	// Whiteout for a path that doesn't exist; should be a no-op, no panic.
	applyWhiteout(entries, "/a/.wh.missing", ".wh.missing")
	if _, ok := entries["/a/b"]; !ok {
		t.Error("unrelated entry should not be affected")
	}
}
