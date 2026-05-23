package cache

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

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
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReadMeta(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")

	archiveStr := "archive.tar.zst"
	size := int64(42)
	digest := "sha256:abc"
	ver := Version{
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
	if !errors.Is(err, ErrNotFound) {
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
