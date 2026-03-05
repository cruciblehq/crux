package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cruciblehq/spec/registry"
)

func TestPathExists(t *testing.T) {
	dir := t.TempDir()

	exists, err := pathExists(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected existing directory to return true")
	}

	exists, err = pathExists(filepath.Join(dir, "nope"))
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected non-existent path to return false")
	}
}

func TestOpenFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	f, err := openFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}

func TestOpenFileNotFound(t *testing.T) {
	_, err := openFile(filepath.Join(t.TempDir(), "missing"))
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")
	content := "hello world"

	digest, size, err := writeFileAtomic(strings.NewReader(content), dir, dest)
	if err != nil {
		t.Fatal(err)
	}

	if size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), size)
	}

	h := sha256.Sum256([]byte(content))
	want := fmt.Sprintf("sha256:%s", hex.EncodeToString(h[:]))
	if digest != want {
		t.Fatalf("expected digest %s, got %s", want, digest)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Fatalf("expected file content %q, got %q", content, data)
	}
}

func TestWriteFileAtomicNoTempLeftover(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")

	if _, _, err := writeFileAtomic(strings.NewReader("x"), dir, dest); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestReadMeta(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")

	archiveStr := "archive.tar.zst"
	size := int64(42)
	digest := "sha256:abc"
	ver := registry.Version{
		Namespace: "ns",
		Resource:  "res",
		String:    "1.0.0",
		Archive:   &archiveStr,
		Size:      &size,
		Digest:    &digest,
		CreatedAt: 1000,
		UpdatedAt: 2000,
	}
	data, _ := json.Marshal(ver)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readMeta(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Namespace != "ns" || got.Resource != "res" || got.String != "1.0.0" {
		t.Fatalf("unexpected version: %+v", got)
	}
	if *got.Digest != digest {
		t.Fatalf("expected digest %q, got %q", digest, *got.Digest)
	}
}

func TestReadMetaNotFound(t *testing.T) {
	_, err := readMeta(filepath.Join(t.TempDir(), "missing.json"))
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReadMetaInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := readMeta(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWriteMeta(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")

	ver, err := writeMeta(path, "ns", "res", "2.0.0", "sha256:def", 99)
	if err != nil {
		t.Fatal(err)
	}

	if ver.Namespace != "ns" || ver.Resource != "res" || ver.String != "2.0.0" {
		t.Fatalf("unexpected version: %+v", ver)
	}
	if *ver.Size != 99 {
		t.Fatalf("expected size 99, got %d", *ver.Size)
	}
	if *ver.Digest != "sha256:def" {
		t.Fatalf("expected digest sha256:def, got %s", *ver.Digest)
	}
	if *ver.Archive != archiveFilename {
		t.Fatalf("expected archive %s, got %s", archiveFilename, *ver.Archive)
	}
	if ver.CreatedAt == 0 || ver.UpdatedAt == 0 {
		t.Fatal("expected non-zero timestamps")
	}

	// Verify the file was written and is valid JSON.
	got, err := readMeta(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Namespace != ver.Namespace || got.String != ver.String {
		t.Fatal("readMeta round-trip mismatch")
	}
}

func TestWriteMetaBadPath(t *testing.T) {
	_, err := writeMeta(filepath.Join(t.TempDir(), "no", "such", "dir", "meta.json"), "ns", "res", "1.0.0", "sha256:abc", 1)
	if err == nil {
		t.Fatal("expected error writing to non-existent directory")
	}
}

func TestPruneEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := pruneEmpty(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("expected empty directory to be removed")
	}
}

func TestPruneEmptyNonEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := pruneEmpty(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir); err != nil {
		t.Fatal("expected non-empty directory to remain")
	}
}

func TestPruneEmptyMissing(t *testing.T) {
	err := pruneEmpty(filepath.Join(t.TempDir(), "gone"))
	if err != nil {
		t.Fatalf("expected nil for missing directory, got %v", err)
	}
}

func TestListSubdirs(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha", "beta"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// Create a regular file that should be excluded.
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	names, err := listSubdirs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 subdirs, got %d: %v", len(names), names)
	}
}

func TestListSubdirsEmpty(t *testing.T) {
	names, err := listSubdirs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 subdirs, got %d", len(names))
	}
}

func TestListSubdirsMissing(t *testing.T) {
	_, err := listSubdirs(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestExtractDirAtomic(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "extracted")

	archive := testArchive(t, map[string]string{"hello.txt": "world"})
	if err := extractDirAtomic(archive, dest); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dest, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "world" {
		t.Fatalf("expected %q, got %q", "world", data)
	}
}

func TestExtractDirAtomicCreatesParent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deep", "nested")

	archive := testArchive(t, map[string]string{"f": "content"})
	if err := extractDirAtomic(archive, dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "f")); err != nil {
		t.Fatal("expected extracted file to exist")
	}
}

func TestExtractDirAtomicBadReader(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out")
	err := extractDirAtomic(bytes.NewReader([]byte("not a valid archive")), dest)
	if err == nil {
		t.Fatal("expected error for invalid archive data")
	}
}
