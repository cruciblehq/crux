package archive

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndExtract(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
	assertFileContent(t, filepath.Join(destDir, "subdir", "nested.txt"), "nested")
	assertDirExists(t, filepath.Join(destDir, "emptydir"))
}

func TestExtractReader(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := ExtractReader(bytes.NewReader(data), destDir); err != nil {
		t.Fatalf("ExtractReader failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
	assertFileContent(t, filepath.Join(destDir, "subdir", "nested.txt"), "nested")
	assertDirExists(t, filepath.Join(destDir, "emptydir"))
}

func TestExtractReaderDestinationExists(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	destDir := t.TempDir()
	err = ExtractReader(bytes.NewReader(data), destDir)

	if err == nil {
		t.Fatal("expected error for existing destination")
	}

	if !errors.Is(err, ErrDestinationExists) {
		t.Fatalf("expected ErrDestinationExists, got: %v", err)
	}
}

func TestExtractReaderInvalidData(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "extracted")

	err := ExtractReader(bytes.NewReader([]byte("not a valid archive")), destDir)
	if err == nil {
		t.Fatal("expected error for invalid data")
	}

	if _, statErr := os.Stat(destDir); statErr == nil {
		t.Fatal("destination should not exist after failed extraction")
	}
}

func TestCreateSymlinkError(t *testing.T) {
	srcDir := t.TempDir()

	target := filepath.Join(srcDir, "target.txt")
	if err := os.WriteFile(target, []byte("target"), 0644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(srcDir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	err := Create(srcDir, archivePath)

	if err == nil {
		t.Fatal("expected error for symlink")
	}

	if !errors.Is(err, ErrSymlink) {
		t.Fatalf("expected ErrSymlink, got: %v", err)
	}

	if _, statErr := os.Stat(archivePath); statErr == nil {
		t.Fatal("archive should be removed on failure")
	}
}

func TestExtractDestinationExists(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	err := Extract(archivePath, destDir)

	if err == nil {
		t.Fatal("expected error for existing destination")
	}

	if !errors.Is(err, ErrDestinationExists) {
		t.Fatalf("expected ErrDestinationExists, got: %v", err)
	}
}

func TestExtractInvalidPath(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatal(err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
}

func TestExtractCleansUpOnFailure(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "nonexistent.tar.zst")
	destDir := filepath.Join(t.TempDir(), "extracted")

	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for nonexistent archive")
	}

	if _, statErr := os.Stat(destDir); statErr == nil {
		t.Fatal("destination should not exist after failed extraction")
	}
}

func TestCreateEmptyDirectory(t *testing.T) {
	srcDir := t.TempDir()

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	assertDirExists(t, destDir)
}

func TestCreateNestedDirectories(t *testing.T) {
	srcDir := t.TempDir()

	nested := filepath.Join(srcDir, "a", "b", "c")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(nested, "deep.txt"), []byte("deep"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "a", "b", "c", "deep.txt"), "deep")
}

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
