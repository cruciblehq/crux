package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
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

func TestWriteArchive(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")
	content := "hello world"

	digest, size, err := writeArchive(strings.NewReader(content), dest)
	if err != nil {
		t.Fatalf("writeArchive() = %v, want nil", err)
	}
	if size != int64(len(content)) {
		t.Errorf("size = %d, want %d", size, len(content))
	}
	h := sha256.Sum256([]byte(content))
	if want := fmt.Sprintf("sha256:%s", hex.EncodeToString(h[:])); digest != want {
		t.Errorf("digest = %s, want %s", digest, want)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", data, content)
	}
}

func TestWriteArchiveReadError(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")

	if _, _, err := writeArchive(iotest.ErrReader(os.ErrClosed), dest); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("writeArchive() = %v, want %v", err, os.ErrClosed)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("destination exists after a failed write, stat = %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has %d entries after a failed write, want 0", len(entries))
	}
}
