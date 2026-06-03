package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathExists(t *testing.T) {
	dir := t.TempDir()

	exists, err := PathExists(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected existing directory to return true")
	}

	exists, err = PathExists(filepath.Join(dir, "nope"))
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected non-existent path to return false")
	}
}

func TestWriteAtomic(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")
	content := "hello world"

	digest, size, err := WriteAtomic(strings.NewReader(content), dir, dest)
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

func TestWriteAtomicNoTempLeftover(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")

	if _, _, err := WriteAtomic(strings.NewReader("x"), dir, dest); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestRemoveDirIfEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := RemoveDirIfEmpty(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("expected empty directory to be removed")
	}
}

func TestRemoveDirIfEmptyNonEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveDirIfEmpty(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir); err != nil {
		t.Fatal("expected non-empty directory to remain")
	}
}

func TestRemoveDirIfEmptyMissing(t *testing.T) {
	err := RemoveDirIfEmpty(filepath.Join(t.TempDir(), "gone"))
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

	names, err := ListSubdirs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 subdirs, got %d: %v", len(names), names)
	}
}

func TestListSubdirsEmpty(t *testing.T) {
	names, err := ListSubdirs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 subdirs, got %d", len(names))
	}
}

func TestListSubdirsMissing(t *testing.T) {
	_, err := ListSubdirs(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}
