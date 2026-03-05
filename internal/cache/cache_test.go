package cache

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func testArchive(t *testing.T, files map[string]string) *bytes.Reader {
	t.Helper()

	var buf bytes.Buffer
	zw, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(zw)

	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buf.Bytes())
}

// Returns the sha256 digest string for the given data.
func testDigest(t *testing.T, data []byte) string {
	t.Helper()
	h := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(h[:]))
}

// Opens a new cache in a temporary directory.
func testOpen(t *testing.T) *Cache {
	t.Helper()
	dir := t.TempDir()
	c, err := OpenAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestOpenAt(t *testing.T) {
	dir := t.TempDir()
	c, err := OpenAt(dir)
	if err != nil {
		t.Fatal(err)
	}

	lockPath := filepath.Join(dir, lockFilename)
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("lock file not created: %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenAtCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "cache")
	c, err := OpenAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("cache directory not created: %v", err)
	}
}

func TestCloseIdempotent(t *testing.T) {
	c := testOpen(t)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	// Second close should be a no-op.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPutAndGet(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"hello.txt": "world"})

	ver, err := c.Put("ns", "res", "1.0.0", archive)
	if err != nil {
		t.Fatal(err)
	}

	if ver.Namespace != "ns" || ver.Resource != "res" || ver.String != "1.0.0" {
		t.Fatalf("unexpected version: %+v", ver)
	}
	if ver.Digest == nil || !strings.HasPrefix(*ver.Digest, "sha256:") {
		t.Fatalf("expected sha256 digest, got %v", ver.Digest)
	}
	if ver.Size == nil || *ver.Size == 0 {
		t.Fatal("expected non-zero size")
	}

	got, err := c.Get("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if got.Namespace != "ns" || got.Resource != "res" || got.String != "1.0.0" {
		t.Fatalf("Get returned wrong version: %+v", got)
	}
	if *got.Digest != *ver.Digest {
		t.Fatalf("digest mismatch: %s != %s", *got.Digest, *ver.Digest)
	}
}

func TestPutReplaces(t *testing.T) {
	c := testOpen(t)

	a1 := testArchive(t, map[string]string{"a.txt": "first"})
	v1, err := c.Put("ns", "res", "1.0.0", a1)
	if err != nil {
		t.Fatal(err)
	}

	a2 := testArchive(t, map[string]string{"b.txt": "second"})
	v2, err := c.Put("ns", "res", "1.0.0", a2)
	if err != nil {
		t.Fatal(err)
	}

	if *v1.Digest == *v2.Digest {
		t.Fatal("expected different digests after replacement")
	}

	got, err := c.Get("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if *got.Digest != *v2.Digest {
		t.Fatal("Get should return the replacement version")
	}
}

func TestPutInvalidatesExtraction(t *testing.T) {
	c := testOpen(t)

	a1 := testArchive(t, map[string]string{"old.txt": "old"})
	if _, err := c.Put("ns", "res", "1.0.0", a1); err != nil {
		t.Fatal(err)
	}

	dir, err := c.Extract("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "old.txt")); err != nil {
		t.Fatal("expected old.txt after first extraction")
	}

	// Replace the archive — extraction should be removed.
	a2 := testArchive(t, map[string]string{"new.txt": "new"})
	if _, err := c.Put("ns", "res", "1.0.0", a2); err != nil {
		t.Fatal(err)
	}

	ok, _ := c.HasExtracted("ns", "res", "1.0.0")
	if ok {
		t.Fatal("expected HasExtracted=false after Put replacement")
	}

	// Re-extract should produce new contents.
	dir2, err := c.Extract("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir2, "new.txt")); err != nil {
		t.Fatal("expected new.txt after re-extraction")
	}
}

func TestPutWithDigestMismatchPrunes(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"f": "data"})

	_, err := c.PutWithDigest("ns", "res", "1.0.0", "sha256:bad", archive)
	if err != ErrDigestMismatch {
		t.Fatalf("expected ErrDigestMismatch, got %v", err)
	}

	// Parent directories should have been pruned.
	nsDir, _ := safeJoin(c.archivesRoot(), "ns")
	if _, err := os.Stat(nsDir); !os.IsNotExist(err) {
		t.Fatal("expected namespace directory to be pruned after digest mismatch")
	}
}

func TestHas(t *testing.T) {
	c := testOpen(t)

	ok, err := c.Has("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected Has=false for missing entry")
	}

	archive := testArchive(t, map[string]string{"f": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	ok, err = c.Has("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected Has=true after Put")
	}
}

func TestGetNotFound(t *testing.T) {
	c := testOpen(t)
	_, err := c.Get("ns", "res", "1.0.0")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestOpenArchive(t *testing.T) {
	c := testOpen(t)

	data := testArchive(t, map[string]string{"test": "content"})
	raw, _ := io.ReadAll(io.NewSectionReader(data, 0, data.Size()))
	data.Seek(0, io.SeekStart)

	if _, err := c.Put("ns", "res", "1.0.0", data); err != nil {
		t.Fatal(err)
	}

	rc, err := c.OpenArchive("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatal("archive content mismatch")
	}
}

func TestOpenArchiveNotFound(t *testing.T) {
	c := testOpen(t)
	_, err := c.OpenArchive("ns", "res", "1.0.0")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPutWithDigest(t *testing.T) {
	c := testOpen(t)

	data := testArchive(t, map[string]string{"f": "data"})
	raw, _ := io.ReadAll(io.NewSectionReader(data, 0, data.Size()))
	data.Seek(0, io.SeekStart)

	expected := testDigest(t, raw)

	ver, err := c.PutWithDigest("ns", "res", "1.0.0", expected, data)
	if err != nil {
		t.Fatal(err)
	}
	if *ver.Digest != expected {
		t.Fatalf("digest mismatch: %s != %s", *ver.Digest, expected)
	}
}

func TestPutWithDigestMismatch(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"f": "data"})

	_, err := c.PutWithDigest("ns", "res", "1.0.0", "sha256:bad", archive)
	if err != ErrDigestMismatch {
		t.Fatalf("expected ErrDigestMismatch, got %v", err)
	}

	// Entry should have been cleaned up.
	ok, _ := c.Has("ns", "res", "1.0.0")
	if ok {
		t.Fatal("entry should have been removed after digest mismatch")
	}
}

func TestDelete(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"f": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	if err := c.Delete("ns", "res", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	ok, _ := c.Has("ns", "res", "1.0.0")
	if ok {
		t.Fatal("entry should be gone after Delete")
	}
}

func TestDeleteIdempotent(t *testing.T) {
	c := testOpen(t)
	// Deleting a non-existent entry should not error.
	if err := c.Delete("ns", "res", "1.0.0"); err != nil {
		t.Fatal(err)
	}
}

func TestDeletePrunesEmptyDirs(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"f": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	if err := c.Delete("ns", "res", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	// The namespace directory should have been pruned.
	nsDir, _ := safeJoin(c.archivesRoot(), "ns")
	if _, err := os.Stat(nsDir); !os.IsNotExist(err) {
		t.Fatal("expected namespace directory to be pruned")
	}
}

func TestDeleteRemovesExtracted(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"hello.txt": "world"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	dir, err := c.Extract("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("extracted dir should exist: %v", err)
	}

	if err := c.Delete("ns", "res", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("extracted directory should be removed by Delete")
	}
}

func TestList(t *testing.T) {
	c := testOpen(t)

	entries := []struct{ ns, res, ver string }{
		{"ns1", "res1", "1.0.0"},
		{"ns1", "res2", "2.0.0"},
		{"ns2", "res1", "0.1.0"},
	}
	for _, e := range entries {
		a := testArchive(t, map[string]string{"f": e.ver})
		if _, err := c.Put(e.ns, e.res, e.ver, a); err != nil {
			t.Fatal(err)
		}
	}

	versions, err := c.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
}

func TestListEmpty(t *testing.T) {
	c := testOpen(t)
	versions, err := c.List()
	if err != nil {
		t.Fatal(err)
	}
	if versions != nil {
		t.Fatalf("expected nil for empty cache, got %v", versions)
	}
}

func TestClear(t *testing.T) {
	c := testOpen(t)

	a := testArchive(t, map[string]string{"f": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", a); err != nil {
		t.Fatal(err)
	}

	// Also extract so we can verify Clear removes both trees.
	if _, err := c.Extract("ns", "res", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	if err := c.Clear(); err != nil {
		t.Fatal(err)
	}

	versions, err := c.List()
	if err != nil {
		t.Fatal(err)
	}
	if versions != nil {
		t.Fatal("expected empty cache after Clear")
	}

	if _, err := os.Stat(c.archivesRoot()); !os.IsNotExist(err) {
		t.Fatal("archives directory should not exist after Clear")
	}
	if _, err := os.Stat(c.extractedRoot()); !os.IsNotExist(err) {
		t.Fatal("extracted directory should not exist after Clear")
	}
}

func TestExtract(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{
		"hello.txt":      "world",
		"sub/nested.txt": "nested content",
	})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	dir, err := c.Extract("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// Verify extracted files exist.
	content, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatalf("expected hello.txt in extracted dir: %v", err)
	}
	if string(content) != "world" {
		t.Fatalf("unexpected content: %q", content)
	}

	nested, err := os.ReadFile(filepath.Join(dir, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("expected sub/nested.txt in extracted dir: %v", err)
	}
	if string(nested) != "nested content" {
		t.Fatalf("unexpected nested content: %q", nested)
	}

	// Verify the extracted path is under the extracted tree.
	if !strings.Contains(dir, extractedDir) {
		t.Fatalf("expected extracted path to contain %q, got %q", extractedDir, dir)
	}
}

func TestExtractIdempotent(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"f.txt": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	dir1, err := c.Extract("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	dir2, err := c.Extract("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	if dir1 != dir2 {
		t.Fatalf("Extract should return same path: %q != %q", dir1, dir2)
	}
}

func TestExtractNotFound(t *testing.T) {
	c := testOpen(t)
	_, err := c.Extract("ns", "res", "1.0.0")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestHasExtracted(t *testing.T) {
	c := testOpen(t)

	ok, err := c.HasExtracted("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected HasExtracted=false before extraction")
	}

	archive := testArchive(t, map[string]string{"f": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	// Still false — Put does not extract.
	ok, err = c.HasExtracted("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected HasExtracted=false before Extract call")
	}

	if _, err := c.Extract("ns", "res", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	ok, err = c.HasExtracted("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected HasExtracted=true after Extract")
	}
}

func TestExtractedInParallelTree(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"f": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	dir, err := c.Extract("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// Extracted path should be under extracted/, not archives/.
	if strings.Contains(dir, archivesDir) {
		t.Fatalf("extracted path should not be under %q: %s", archivesDir, dir)
	}
	if !strings.Contains(dir, extractedDir) {
		t.Fatalf("extracted path should be under %q: %s", extractedDir, dir)
	}
}

func TestDirectoryLayout(t *testing.T) {
	c := testOpen(t)

	archive := testArchive(t, map[string]string{"f": "data"})
	if _, err := c.Put("ns", "res", "1.0.0", archive); err != nil {
		t.Fatal(err)
	}

	// Verify archive tree structure.
	metPath, err := c.metaPath("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(metPath); err != nil {
		t.Fatalf("meta.json should exist: %v", err)
	}

	archPath, err := c.archivePath("ns", "res", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(archPath); err != nil {
		t.Fatalf("archive.tar.zst should exist: %v", err)
	}

	// Verify paths are structured correctly.
	vDir, _ := c.versionDir("ns", "res", "1.0.0")
	expected := filepath.Join(c.root, archivesDir, "ns", "res", "1.0.0")
	if vDir != expected {
		t.Fatalf("unexpected versionDir: %s", vDir)
	}

	eDir, _ := c.extractedVersionDir("ns", "res", "1.0.0")
	expectedExtracted := filepath.Join(c.root, extractedDir, "ns", "res", "1.0.0")
	if eDir != expectedExtracted {
		t.Fatalf("unexpected extractedVersionDir: %s", eDir)
	}
}

func TestMultipleNamespacesAndResources(t *testing.T) {
	c := testOpen(t)

	entries := []struct{ ns, res, ver string }{
		{"alpha", "widget", "1.0.0"},
		{"alpha", "widget", "2.0.0"},
		{"alpha", "service", "1.0.0"},
		{"beta", "widget", "1.0.0"},
	}

	for _, e := range entries {
		a := testArchive(t, map[string]string{"id": e.ns + "/" + e.res + "@" + e.ver})
		if _, err := c.Put(e.ns, e.res, e.ver, a); err != nil {
			t.Fatal(err)
		}
	}

	versions, err := c.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 4 {
		t.Fatalf("expected 4 versions, got %d", len(versions))
	}

	// Delete one, verify the rest are intact.
	if err := c.Delete("alpha", "widget", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	versions, err = c.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions after Delete, got %d", len(versions))
	}

	// The resource dir should still exist (widget 2.0.0 remains).
	rDir, _ := safeJoin(c.archivesRoot(), "alpha", "widget")
	if _, err := os.Stat(rDir); err != nil {
		t.Fatal("resource dir should still exist for remaining version")
	}
}

func TestDeletePrunesOnlyEmptyDirs(t *testing.T) {
	c := testOpen(t)

	// Two versions under the same namespace/resource.
	for _, ver := range []string{"1.0.0", "2.0.0"} {
		a := testArchive(t, map[string]string{"f": ver})
		if _, err := c.Put("ns", "res", ver, a); err != nil {
			t.Fatal(err)
		}
	}

	if err := c.Delete("ns", "res", "1.0.0"); err != nil {
		t.Fatal(err)
	}

	// Resource dir should still exist because 2.0.0 remains.
	rDir, _ := safeJoin(c.archivesRoot(), "ns", "res")
	if _, err := os.Stat(rDir); err != nil {
		t.Fatal("resource dir should not be pruned while versions remain")
	}

	// Now delete the last one.
	if err := c.Delete("ns", "res", "2.0.0"); err != nil {
		t.Fatal(err)
	}

	// Now the whole tree should be pruned.
	nsDir, _ := safeJoin(c.archivesRoot(), "ns")
	if _, err := os.Stat(nsDir); !os.IsNotExist(err) {
		t.Fatal("namespace dir should be pruned when empty")
	}
}

func TestValidatePathComponent(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "my-resource", false},
		{"valid with dots", "1.0.0", false},
		{"empty", "", true},
		{"dot", ".", true},
		{"dotdot", "..", true},
		{"slash", "ns/evil", true},
		{"backslash", "res\\evil", true},
		{"path traversal", "../../etc", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePathComponent(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validatePathComponent(%q) error = %v, wantErr = %v",
					tc.input, err, tc.wantErr)
			}
		})
	}
}

func TestValidationRejectsBeforeLock(t *testing.T) {
	c := testOpen(t)

	_, err := c.Has("", "res", "1.0.0")
	if err == nil {
		t.Fatal("expected validation error for empty namespace")
	}

	_, err = c.Get("ns", "..", "1.0.0")
	if err == nil {
		t.Fatal("expected validation error for dotdot resource")
	}

	_, err = c.OpenArchive("ns/bad", "res", "1.0.0")
	if err == nil {
		t.Fatal("expected validation error for slash in namespace")
	}

	_, err = c.HasExtracted("ns", "res", "")
	if err == nil {
		t.Fatal("expected validation error for empty version")
	}

	_, err = c.Extract(".", "res", "1.0.0")
	if err == nil {
		t.Fatal("expected validation error for dot namespace")
	}

	_, err = c.Put("ns", "res\\bad", "1.0.0", strings.NewReader("data"))
	if err == nil {
		t.Fatal("expected validation error for backslash in resource")
	}

	_, err = c.PutWithDigest("..", "res", "1.0.0", "sha256:abc", strings.NewReader("data"))
	if err == nil {
		t.Fatal("expected validation error for dotdot namespace")
	}

	err = c.Delete("ns", "res", "../../etc")
	if err == nil {
		t.Fatal("expected validation error for path traversal in version")
	}
}
