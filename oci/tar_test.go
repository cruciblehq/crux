package oci

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Reads all entries from a tar archive, returning a map of entry names to their
// content (or "<dir>" for directories). Fails the test if any entry has a
// non-zero modification time.
func readTarEntries(t *testing.T, data []byte) map[string]string {
	t.Helper()
	tr := tar.NewReader(bytes.NewReader(data))
	entries := make(map[string]string)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading tar: %v", err)
		}

		if header.Typeflag == tar.TypeReg {
			content, _ := io.ReadAll(tr)
			entries[header.Name] = string(content)
		} else {
			entries[header.Name] = "<dir>"
		}

		if header.ModTime.Unix() != 0 {
			t.Errorf("ModTime should be Unix epoch for %s, got %v", header.Name, header.ModTime)
		}
	}
	return entries
}

func TestCreateTarFromDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directory structure
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := createTarFromDir(srcDir, "/app")
	if err != nil {
		t.Fatalf("createTarFromDir: %v", err)
	}

	entries := readTarEntries(t, data)

	if entries["/app/file1.txt"] != "content1" {
		t.Errorf("expected file1.txt content 'content1', got %q", entries["/app/file1.txt"])
	}
	if entries["/app/subdir/file2.txt"] != "content2" {
		t.Errorf("expected file2.txt content 'content2', got %q", entries["/app/subdir/file2.txt"])
	}
}

func TestCreateTarFromBytes(t *testing.T) {
	content := []byte("hello world")
	data, err := createTarFromBytes(content, "/bin/app", 0755)
	if err != nil {
		t.Fatalf("createTarFromBytes: %v", err)
	}

	tr := tar.NewReader(bytes.NewReader(data))
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("reading tar: %v", err)
	}

	if header.Name != "/bin/app" {
		t.Errorf("expected name '/bin/app', got %q", header.Name)
	}
	if header.Mode != 0755 {
		t.Errorf("expected mode 0755, got %o", header.Mode)
	}
	if header.ModTime.Unix() != 0 {
		t.Errorf("ModTime should be Unix epoch, got %v", header.ModTime)
	}

	got, _ := io.ReadAll(tr)
	if string(got) != "hello world" {
		t.Errorf("expected content 'hello world', got %q", string(got))
	}

	// Should be only one entry
	_, err = tr.Next()
	if err != io.EOF {
		t.Error("expected only one entry in tar")
	}
}

func TestExtractTar(t *testing.T) {
	// Create a tar archive in memory
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files := []struct {
		name    string
		content string
		isDir   bool
	}{
		{"dir", "", true},
		{"dir/file.txt", "file content", false},
		{"root.txt", "root content", false},
	}

	for _, f := range files {
		if f.isDir {
			tw.WriteHeader(&tar.Header{
				Name:     f.name,
				Mode:     0755,
				Typeflag: tar.TypeDir,
			})
		} else {
			tw.WriteHeader(&tar.Header{
				Name:     f.name,
				Mode:     0644,
				Size:     int64(len(f.content)),
				Typeflag: tar.TypeReg,
			})
			tw.Write([]byte(f.content))
		}
	}
	tw.Close()

	// Extract
	destDir := t.TempDir()
	if err := extractTar(&buf, destDir); err != nil {
		t.Fatalf("extractTar: %v", err)
	}

	// Verify
	content, err := os.ReadFile(filepath.Join(destDir, "dir", "file.txt"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(content) != "file content" {
		t.Errorf("expected 'file content', got %q", string(content))
	}

	content, err = os.ReadFile(filepath.Join(destDir, "root.txt"))
	if err != nil {
		t.Fatalf("reading extracted root.txt: %v", err)
	}
	if string(content) != "root content" {
		t.Errorf("expected 'root content', got %q", string(content))
	}
}

func TestExtractTar_RejectsPathTraversal(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	tw.WriteHeader(&tar.Header{
		Name:     "../escape.txt",
		Mode:     0644,
		Size:     4,
		Typeflag: tar.TypeReg,
	})
	tw.Write([]byte("evil"))
	tw.Close()

	destDir := t.TempDir()
	err := extractTar(&buf, destDir)
	if err == nil {
		t.Error("expected error for path traversal attempt")
	}
}

func TestExtractTar_RejectsAbsolutePath(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	tw.WriteHeader(&tar.Header{
		Name:     "/etc/passwd",
		Mode:     0644,
		Size:     4,
		Typeflag: tar.TypeReg,
	})
	tw.Write([]byte("evil"))
	tw.Close()

	destDir := t.TempDir()
	err := extractTar(&buf, destDir)
	if err == nil {
		t.Error("expected error for absolute path")
	}
}

func TestWriteDirToTarball(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write to tarball
	tarPath := filepath.Join(tmpDir, "output.tar")
	if err := writeDirToTarball(srcDir, tarPath); err != nil {
		t.Fatalf("writeDirToTarball: %v", err)
	}

	// Read and verify
	f, err := os.Open(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tr := tar.NewReader(f)
	found := false
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading tar: %v", err)
		}
		if header.Name == "test.txt" {
			found = true
			content, _ := io.ReadAll(tr)
			if string(content) != "test" {
				t.Errorf("expected content 'test', got %q", string(content))
			}
		}
	}
	if !found {
		t.Error("test.txt not found in tarball")
	}
}

func TestWriteTarEntry(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := writeTarEntry(tw, testFile, "custom/path.txt", info, preserveModTime); err != nil {
		t.Fatalf("writeTarEntry: %v", err)
	}
	tw.Close()

	tr := tar.NewReader(&buf)
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("reading tar: %v", err)
	}

	if header.Name != "custom/path.txt" {
		t.Errorf("expected name 'custom/path.txt', got %q", header.Name)
	}

	content, _ := io.ReadAll(tr)
	if string(content) != "content" {
		t.Errorf("expected content 'content', got %q", string(content))
	}
}

func TestWriteTarEntry_ZerosModTime(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set a specific modification time
	pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Chtimes(testFile, pastTime, pastTime)

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := writeTarEntry(tw, testFile, "file.txt", info, zeroModTime); err != nil {
		t.Fatalf("writeTarEntry: %v", err)
	}
	tw.Close()

	tr := tar.NewReader(&buf)
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("reading tar: %v", err)
	}

	if header.ModTime.Unix() != 0 {
		t.Errorf("expected Unix epoch ModTime, got %v", header.ModTime)
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "output.txt")

	content := bytes.NewReader([]byte("hello world"))
	if err := writeFile(destPath, content); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(got))
	}
}

func TestRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory with content
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested", "deep"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "mid.txt"), []byte("mid"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "deep", "leaf.txt"), []byte("leaf"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create tar
	data, err := createTarFromDir(srcDir, "")
	if err != nil {
		t.Fatalf("createTarFromDir: %v", err)
	}

	// Extract to new location
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := extractTar(bytes.NewReader(data), destDir); err != nil {
		t.Fatalf("extractTar: %v", err)
	}

	// Verify all files
	tests := []struct {
		path    string
		content string
	}{
		{"root.txt", "root"},
		{"nested/mid.txt", "mid"},
		{"nested/deep/leaf.txt", "leaf"},
	}

	for _, tt := range tests {
		got, err := os.ReadFile(filepath.Join(destDir, tt.path))
		if err != nil {
			t.Errorf("reading %s: %v", tt.path, err)
			continue
		}
		if string(got) != tt.content {
			t.Errorf("%s: expected %q, got %q", tt.path, tt.content, string(got))
		}
	}
}
