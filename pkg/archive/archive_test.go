package archive

import (
	"archive/tar"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
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

func TestExtractFromReader(t *testing.T) {
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
	if err := ExtractFromReader(bytes.NewReader(data), destDir); err != nil {
		t.Fatalf("ExtractFromReader failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
	assertFileContent(t, filepath.Join(destDir, "subdir", "nested.txt"), "nested")
	assertDirExists(t, filepath.Join(destDir, "emptydir"))
}

func TestExtractFromReaderDestinationExists(t *testing.T) {
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
	err = ExtractFromReader(bytes.NewReader(data), destDir)

	if err == nil {
		t.Fatal("expected error for existing destination")
	}
}

func TestExtractFromReaderInvalidData(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "extracted")

	err := ExtractFromReader(bytes.NewReader([]byte("not a valid archive")), destDir)
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

	if !errors.Is(err, ErrUnsupportedFileType) {
		t.Fatalf("expected ErrUnsupportedFileType, got: %v", err)
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

	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected os.ErrExist, got: %v", err)
	}
}

func TestExtractPathTraversal(t *testing.T) {
	// Manually craft a malicious archive with path traversal
	destDir := filepath.Join(t.TempDir(), "extracted")
	maliciousArchive := createMaliciousArchive(t, "../etc/passwd")

	err := ExtractFromReader(maliciousArchive, destDir)
	if err == nil {
		t.Fatal("expected error for path traversal attempt")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got: %v", err)
	}
}

func TestExtractAbsolutePath(t *testing.T) {
	// Manually craft an archive with absolute path
	destDir := filepath.Join(t.TempDir(), "extracted")
	maliciousArchive := createMaliciousArchive(t, "/etc/passwd")

	err := ExtractFromReader(maliciousArchive, destDir)
	if err == nil {
		t.Fatal("expected error for absolute path")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got: %v", err)
	}
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

func createMaliciousArchive(t *testing.T, maliciousPath string) *bytes.Reader {
	t.Helper()

	var buf bytes.Buffer
	zw, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}

	tw := tar.NewWriter(zw)

	// Create a tar entry with malicious path
	header := &tar.Header{
		Name:     maliciousPath,
		Mode:     0644,
		Size:     5,
		Typeflag: tar.TypeReg,
	}

	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}

	if _, err := tw.Write([]byte("pwned")); err != nil {
		t.Fatal(err)
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	return bytes.NewReader(buf.Bytes())
}
