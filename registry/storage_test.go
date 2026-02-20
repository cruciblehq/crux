package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cruciblehq/crux/archive"
)

func TestArchiveDirectoryPath(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	tests := []struct {
		name      string
		namespace string
		resource  string
		version   string
		want      string
	}{
		{
			name:      "simple path",
			namespace: "test-ns",
			resource:  "test-resource",
			version:   "1.0.0",
			want:      filepath.Join(tempDir, "test-ns", "test-resource", "1.0.0"),
		},
		{
			name:      "nested namespace",
			namespace: "my-namespace",
			resource:  "my-app",
			version:   "2.1.3",
			want:      filepath.Join(tempDir, "my-namespace", "my-app", "2.1.3"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.archiveDirectoryPath(tt.namespace, tt.resource, tt.version)
			if got != tt.want {
				t.Errorf("archiveDirectoryPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArchiveTempPath(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	tests := []struct {
		name      string
		namespace string
		resource  string
		version   string
		want      string
	}{
		{
			name:      "temp path",
			namespace: "test-ns",
			resource:  "test-resource",
			version:   "1.0.0",
			want:      filepath.Join(tempDir, "test-ns", "test-resource", "1.0.0", TemporaryUploadSuffix),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.archiveTempPath(tt.namespace, tt.resource, tt.version)
			if got != tt.want {
				t.Errorf("archiveTempPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArchiveFinalPath(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	tests := []struct {
		name      string
		namespace string
		resource  string
		version   string
		digest    string
		want      string
	}{
		{
			name:      "final path",
			namespace: "test-ns",
			resource:  "test-resource",
			version:   "1.0.0",
			digest:    "abc123",
			want:      filepath.Join(tempDir, "test-ns", "test-resource", "1.0.0", "abc123"+archive.Zstd.String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.archiveFinalPath(tt.namespace, tt.resource, tt.version, tt.digest)
			if got != tt.want {
				t.Errorf("archiveFinalPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStoreArchiveFile_Success(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	// Create test data
	testData := []byte("test archive content for hashing")
	reader := bytes.NewReader(testData)

	// Calculate expected digest
	hasher := sha256.New()
	hasher.Write(testData)
	expectedDigest := hex.EncodeToString(hasher.Sum(nil))

	// Store the archive
	digest, path, size, err := registry.storeArchiveFile("test-ns", "test-resource", "1.0.0", reader)

	// Verify no error
	if err != nil {
		t.Fatalf("storeArchiveFile() error = %v", err)
	}

	// Verify digest
	if digest != expectedDigest {
		t.Errorf("digest = %q, want %q", digest, expectedDigest)
	}

	// Verify size
	if size != int64(len(testData)) {
		t.Errorf("size = %d, want %d", size, len(testData))
	}

	// Verify path
	expectedPath := filepath.Join(tempDir, "test-ns", "test-resource", "1.0.0", expectedDigest+archive.Zstd.String())
	if path != expectedPath {
		t.Errorf("path = %q, want %q", path, expectedPath)
	}

	// Verify file exists and contains correct data
	storedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read stored file: %v", err)
	}
	if !bytes.Equal(storedData, testData) {
		t.Errorf("stored data does not match input data")
	}

	// Verify temp file was cleaned up
	tempPath := filepath.Join(tempDir, "test-ns", "test-resource", "1.0.0", TemporaryUploadSuffix)
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Errorf("temporary file still exists at %q", tempPath)
	}
}

func TestStoreArchiveFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	// Create large test data (1 MB)
	testData := bytes.Repeat([]byte("a"), 1024*1024)
	reader := bytes.NewReader(testData)

	// Calculate expected digest
	hasher := sha256.New()
	hasher.Write(testData)
	expectedDigest := hex.EncodeToString(hasher.Sum(nil))

	// Store the archive
	digest, path, size, err := registry.storeArchiveFile("test-ns", "test-resource", "2.0.0", reader)

	// Verify no error
	if err != nil {
		t.Fatalf("storeArchiveFile() error = %v", err)
	}

	// Verify digest
	if digest != expectedDigest {
		t.Errorf("digest = %q, want %q", digest, expectedDigest)
	}

	// Verify size
	if size != int64(len(testData)) {
		t.Errorf("size = %d, want %d", size, len(testData))
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("stored file does not exist: %v", err)
	}
}

func TestStoreArchiveFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	// Create empty reader
	reader := bytes.NewReader([]byte{})

	// Calculate expected digest for empty data
	hasher := sha256.New()
	expectedDigest := hex.EncodeToString(hasher.Sum(nil))

	// Store the archive
	digest, path, size, err := registry.storeArchiveFile("test-ns", "test-resource", "0.0.1", reader)

	// Verify no error
	if err != nil {
		t.Fatalf("storeArchiveFile() error = %v", err)
	}

	// Verify digest
	if digest != expectedDigest {
		t.Errorf("digest = %q, want %q", digest, expectedDigest)
	}

	// Verify size
	if size != 0 {
		t.Errorf("size = %d, want 0", size)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("stored file does not exist: %v", err)
	}
}

func TestStoreArchiveFile_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	// Ensure directory doesn't exist yet
	archiveDir := filepath.Join(tempDir, "new-ns", "new-resource", "1.0.0")
	if _, err := os.Stat(archiveDir); !os.IsNotExist(err) {
		t.Fatalf("archive directory should not exist yet")
	}

	// Store archive
	testData := []byte("test")
	_, _, _, err := registry.storeArchiveFile("new-ns", "new-resource", "1.0.0", bytes.NewReader(testData))

	// Verify no error
	if err != nil {
		t.Fatalf("storeArchiveFile() error = %v", err)
	}

	// Verify directory was created
	if stat, err := os.Stat(archiveDir); err != nil {
		t.Errorf("archive directory was not created: %v", err)
	} else if !stat.IsDir() {
		t.Errorf("archive path is not a directory")
	}
}

func TestStoreArchiveFile_ReadError(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	// Create a reader that returns an error
	errorReader := &errorReader{err: io.ErrUnexpectedEOF}

	// Attempt to store archive
	_, _, _, err := registry.storeArchiveFile("test-ns", "test-resource", "1.0.0", errorReader)

	// Verify error occurred
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify error is the expected one
	if err != io.ErrUnexpectedEOF {
		t.Errorf("error = %v, want %v", err, io.ErrUnexpectedEOF)
	}

	// Verify temp file was cleaned up
	tempPath := filepath.Join(tempDir, "test-ns", "test-resource", "1.0.0", TemporaryUploadSuffix)
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Errorf("temporary file was not cleaned up")
	}
}

func TestStoreArchiveFile_NestedDirectories(t *testing.T) {
	tempDir := t.TempDir()
	registry := &SQLRegistry{archiveRoot: tempDir}

	// Test with deeply nested structure
	testData := []byte("nested test")
	_, path, _, err := registry.storeArchiveFile("org-ns", "my-resource", "10.20.30", bytes.NewReader(testData))

	if err != nil {
		t.Fatalf("storeArchiveFile() error = %v", err)
	}

	// Verify path contains all nested components
	if !strings.Contains(path, "org-ns") {
		t.Errorf("path missing namespace component")
	}
	if !strings.Contains(path, "my-resource") {
		t.Errorf("path missing resource component")
	}
	if !strings.Contains(path, "10.20.30") {
		t.Errorf("path missing version component")
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("stored file does not exist: %v", err)
	}
}

// Mock reader that returns an error
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
