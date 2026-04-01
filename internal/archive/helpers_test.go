package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestFiles(t *testing.T, dir string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	emptydir := filepath.Join(dir, "emptydir")
	if err := os.MkdirAll(emptydir, 0755); err != nil {
		t.Fatal(err)
	}
}

func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	if string(content) != expected {
		t.Fatalf("expected %q, got %q", expected, string(content))
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory %s does not exist: %v", path, err)
	}

	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}
