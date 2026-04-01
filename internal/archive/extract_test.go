package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

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
	if err := ExtractFromReader(bytes.NewReader(data), destDir, Zstd); err != nil {
		t.Fatalf("ExtractFromReader failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
	assertFileContent(t, filepath.Join(destDir, "subdir", "nested.txt"), "nested")
	assertDirExists(t, filepath.Join(destDir, "emptydir"))
}

func TestExtractFromReaderIntoExistingDirectory(t *testing.T) {
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
	if err := ExtractFromReader(bytes.NewReader(data), destDir, Zstd); err != nil {
		t.Fatalf("ExtractFromReader into existing dir failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
}

func TestExtractFromReaderInvalidData(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "extracted")

	err := ExtractFromReader(bytes.NewReader([]byte("not a valid archive")), destDir, Zstd)
	if err == nil {
		t.Fatal("expected error for invalid data")
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
	maliciousArchive := createMaliciousZstdArchive(t, "../etc/passwd")

	err := ExtractFromReader(maliciousArchive, destDir, Zstd)
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
	maliciousArchive := createMaliciousZstdArchive(t, "/etc/passwd")

	err := ExtractFromReader(maliciousArchive, destDir, Zstd)
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

func TestExtractFilePermissions(t *testing.T) {
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

	// Check file permissions (should match source: 0644)
	info, err := os.Stat(filepath.Join(destDir, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("file mode = %o, want %o", info.Mode().Perm(), 0644)
	}

	// Check directory permissions
	info, err = os.Stat(filepath.Join(destDir, "subdir"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("dir mode = %o, want %o", info.Mode().Perm(), 0755)
	}
}

func TestExtractCleansUpOnMaliciousEntry(t *testing.T) {
	// Archive with a valid file followed by a symlink entry
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "good.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "malicious.tar.zst")

	// Build the archive manually to include a symlink
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	tw.WriteHeader(&tar.Header{
		Name:     "good.txt",
		Size:     5,
		Mode:     0644,
		Typeflag: tar.TypeReg,
	})
	tw.Write([]byte("hello"))

	tw.WriteHeader(&tar.Header{
		Name:     "evil",
		Typeflag: tar.TypeSymlink,
		Linkname: "/etc/passwd",
	})

	tw.Close()
	zw.Close()

	if err := os.WriteFile(archivePath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for symlink entry")
	}

	// Destination should be cleaned up by Extract
	if _, statErr := os.Stat(destDir); statErr == nil {
		t.Fatal("destination should not exist after failed extraction")
	}
}

func TestExtractPreservesFileMode(t *testing.T) {
	// Create an archive with files of different modes
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	script := []byte("#!/bin/sh\necho hello")
	tw.WriteHeader(&tar.Header{
		Name:     "script.sh",
		Size:     int64(len(script)),
		Mode:     0755,
		Typeflag: tar.TypeReg,
	})
	tw.Write(script)

	data := []byte("plain data")
	tw.WriteHeader(&tar.Header{
		Name:     "data.txt",
		Size:     int64(len(data)),
		Mode:     0644,
		Typeflag: tar.TypeReg,
	})
	tw.Write(data)

	tw.Close()
	zw.Close()

	dest := filepath.Join(t.TempDir(), "preserved")
	if err := ExtractFromReader(bytes.NewReader(buf.Bytes()), dest, Zstd); err != nil {
		t.Fatalf("ExtractFromReader: %v", err)
	}

	info, _ := os.Stat(filepath.Join(dest, "script.sh"))
	if info.Mode().Perm() != 0755 {
		t.Errorf("script.sh mode = %o, want 0755", info.Mode().Perm())
	}

	info, _ = os.Stat(filepath.Join(dest, "data.txt"))
	if info.Mode().Perm() != 0644 {
		t.Errorf("data.txt mode = %o, want 0644", info.Mode().Perm())
	}
}

func TestExtractFromReaderGzip(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	archivePath := filepath.Join(t.TempDir(), "test.tar.gz")
	if err := Create(srcDir, archivePath); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := ExtractFromReader(bytes.NewReader(data), destDir, Gzip); err != nil {
		t.Fatalf("ExtractFromReader failed: %v", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
}

func TestExtractUnsupportedFormat(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "test.zip")
	if err := os.WriteFile(archivePath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(t.TempDir(), "extracted")
	err := Extract(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("expected ErrUnsupportedFormat, got: %v", err)
	}
}

func TestExtractGzipPathTraversal(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "extracted")
	maliciousArchive := createMaliciousGzipArchive(t, "../etc/passwd")

	err := ExtractFromReader(maliciousArchive, destDir, Gzip)
	if err == nil {
		t.Fatal("expected error for path traversal attempt")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got: %v", err)
	}
}

func TestExtractSkipsPAXHeaders(t *testing.T) {
	// Archive with a PAX extended header followed by a regular file.
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	// PAX extended header entry.
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeXHeader,
		Name:     "PaxHeaders.0/file.txt",
		Size:     0,
	})

	// PAX global header entry.
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeXGlobalHeader,
		Name:     "GlobalHead.0",
		Size:     0,
	})

	content := []byte("hello")
	tw.WriteHeader(&tar.Header{
		Name:     "file.txt",
		Size:     int64(len(content)),
		Mode:     0644,
		Typeflag: tar.TypeReg,
	})
	tw.Write(content)

	tw.Close()
	zw.Close()

	dest := filepath.Join(t.TempDir(), "extracted")
	if err := ExtractFromReader(bytes.NewReader(buf.Bytes()), dest, Zstd); err != nil {
		t.Fatalf("ExtractFromReader: %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "file.txt"), "hello")
}

func TestExtractHardLink(t *testing.T) {
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	// Write the original file.
	content := []byte("linked content")
	tw.WriteHeader(&tar.Header{
		Name:     "original.txt",
		Size:     int64(len(content)),
		Mode:     0644,
		Typeflag: tar.TypeReg,
	})
	tw.Write(content)

	// Write a hard link to the original.
	tw.WriteHeader(&tar.Header{
		Name:     "subdir/link.txt",
		Typeflag: tar.TypeLink,
		Linkname: "original.txt",
	})

	tw.Close()
	zw.Close()

	dest := filepath.Join(t.TempDir(), "extracted")
	if err := ExtractFromReader(bytes.NewReader(buf.Bytes()), dest, Zstd); err != nil {
		t.Fatalf("ExtractFromReader: %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "original.txt"), "linked content")
	assertFileContent(t, filepath.Join(dest, "subdir", "link.txt"), "linked content")
}

func TestExtractHardLinkWithDotSlashPrefix(t *testing.T) {
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	content := []byte("gnu-style content")
	tw.WriteHeader(&tar.Header{
		Name:     "./original.txt",
		Size:     int64(len(content)),
		Mode:     0644,
		Typeflag: tar.TypeReg,
	})
	tw.Write(content)

	// Hard link with ./ prefix on both name and linkname.
	tw.WriteHeader(&tar.Header{
		Name:     "./link.txt",
		Typeflag: tar.TypeLink,
		Linkname: "./original.txt",
	})

	tw.Close()
	zw.Close()

	dest := filepath.Join(t.TempDir(), "extracted")
	if err := ExtractFromReader(bytes.NewReader(buf.Bytes()), dest, Zstd); err != nil {
		t.Fatalf("ExtractFromReader: %v", err)
	}

	assertFileContent(t, filepath.Join(dest, "original.txt"), "gnu-style content")
	assertFileContent(t, filepath.Join(dest, "link.txt"), "gnu-style content")
}

func TestExtractHardLinkEscape(t *testing.T) {
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	// Hard link targeting a path outside the destination.
	tw.WriteHeader(&tar.Header{
		Name:     "evil.txt",
		Typeflag: tar.TypeLink,
		Linkname: "../etc/passwd",
	})

	tw.Close()
	zw.Close()

	dest := filepath.Join(t.TempDir(), "extracted")
	err := ExtractFromReader(bytes.NewReader(buf.Bytes()), dest, Zstd)
	if err == nil {
		t.Fatal("expected error for hard link escape")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got: %v", err)
	}
}

func TestExtractHardLinkAbsolutePath(t *testing.T) {
	var buf bytes.Buffer
	zw, _ := zstd.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	// Hard link with an absolute linkname.
	tw.WriteHeader(&tar.Header{
		Name:     "evil.txt",
		Typeflag: tar.TypeLink,
		Linkname: "/etc/passwd",
	})

	tw.Close()
	zw.Close()

	dest := filepath.Join(t.TempDir(), "extracted")
	err := ExtractFromReader(bytes.NewReader(buf.Bytes()), dest, Zstd)
	if err == nil {
		t.Fatal("expected error for hard link with absolute path")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got: %v", err)
	}
}

func createMaliciousZstdArchive(t *testing.T, maliciousPath string) *bytes.Reader {
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

func createMaliciousGzipArchive(t *testing.T, maliciousPath string) *bytes.Reader {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

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

	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	return bytes.NewReader(buf.Bytes())
}
