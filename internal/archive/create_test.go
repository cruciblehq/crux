package archive

import (
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

	if !errors.Is(err, ErrUnsupportedFileType) {
		t.Fatalf("expected ErrUnsupportedFileType, got: %v", err)
	}

	if _, statErr := os.Stat(archivePath); statErr == nil {
		t.Fatal("archive should be removed on failure")
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

func TestCreateNonExistentSource(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "test.tar.zst")
	err := Create("/nonexistent/path/that/does/not/exist", archivePath)
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
	if !errors.Is(err, ErrCreateFailed) {
		t.Fatalf("expected ErrCreateFailed, got: %v", err)
	}
}

func TestCreateNonExistentDestDir(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "deep", "nested", "test.tar.zst")
	err := Create(srcDir, archivePath)
	if err == nil {
		t.Fatal("expected error when dest directory doesn't exist")
	}
	if !errors.Is(err, ErrCreateFailed) {
		t.Fatalf("expected ErrCreateFailed, got: %v", err)
	}
}

func TestCreateAndExtractGzip(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	archivePath := filepath.Join(t.TempDir(), "test.tar.gz")
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

func TestCreateAndExtractTgz(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	archivePath := filepath.Join(t.TempDir(), "test.tgz")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Extract using .tgz extension directly
	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(archivePath, destDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
}

func TestCreateUnsupportedFormat(t *testing.T) {
	srcDir := t.TempDir()
	archivePath := filepath.Join(t.TempDir(), "test.zip")
	err := Create(srcDir, archivePath)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("expected ErrUnsupportedFormat, got: %v", err)
	}
}
