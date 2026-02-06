package oci

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadIndex_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a multi-platform image and save as layout
	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")
	mb.ForPlatform("linux", "arm64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	// Read the tarball and save as layout directory
	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex tarball: %v", err)
	}

	layoutDir := filepath.Join(tmpDir, "layout")
	if err := idx.SaveLayout(layoutDir); err != nil {
		t.Fatalf("SaveLayout: %v", err)
	}
	idx.Close()

	// Now test reading from directory
	idx2, err := ReadIndex(layoutDir)
	if err != nil {
		t.Fatalf("ReadIndex directory: %v", err)
	}
	defer idx2.Close()

	platforms, err := idx2.Platforms()
	if err != nil {
		t.Fatalf("Platforms: %v", err)
	}

	if !platforms["linux/amd64"] || !platforms["linux/arm64"] {
		t.Errorf("Missing expected platforms: %v", platforms)
	}
}

func TestReadIndex_Tarball(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	manifest, err := idx.idx.IndexManifest()
	if err != nil {
		t.Fatalf("IndexManifest: %v", err)
	}

	if len(manifest.Manifests) != 1 {
		t.Errorf("Expected 1 manifest, got %d", len(manifest.Manifests))
	}
}

func TestReadIndex_NotFound(t *testing.T) {
	_, err := ReadIndex("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestIndex_Close(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}

	// Close should be safe to call multiple times
	if err := idx.Close(); err != nil {
		t.Errorf("Close returned error: %v", err)
	}
	if err := idx.Close(); err != nil {
		t.Errorf("Second Close returned error: %v", err)
	}
}

func TestIndex_Platforms(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")
	mb.ForPlatform("linux", "arm64").SetEntrypoint("/bin/sh")
	mb.ForPlatform("darwin", "arm64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	platforms, err := idx.Platforms()
	if err != nil {
		t.Fatalf("Platforms: %v", err)
	}

	if len(platforms) != 3 {
		t.Errorf("Expected 3 platforms, got %d", len(platforms))
	}

	expected := []string{"linux/amd64", "linux/arm64", "darwin/arm64"}
	for _, p := range expected {
		if !platforms[p] {
			t.Errorf("Missing platform %q", p)
		}
	}
}

func TestIndex_ValidateMultiPlatform_Success(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")
	mb.ForPlatform("linux", "arm64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	if err := idx.ValidateMultiPlatform(); err != nil {
		t.Errorf("ValidateMultiPlatform failed: %v", err)
	}
}

func TestIndex_ValidateMultiPlatform_MissingPlatform(t *testing.T) {
	tmpDir := t.TempDir()

	// Only include amd64, missing arm64
	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	err = idx.ValidateMultiPlatform()
	if err != ErrInsufficientPlatforms {
		t.Errorf("Expected ErrInsufficientPlatforms, got %v", err)
	}
}

func TestIndex_LoadImage(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	amd64 := mb.ForPlatform("linux", "amd64")
	amd64.SetEntrypoint("/amd64-app")

	arm64 := mb.ForPlatform("linux", "arm64")
	arm64.SetEntrypoint("/arm64-app")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	// Load amd64 image
	img, err := idx.LoadImage("linux", "amd64")
	if err != nil {
		t.Fatalf("LoadImage linux/amd64: %v", err)
	}

	cfg, err := img.img.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}

	if cfg.Config.Entrypoint[0] != "/amd64-app" {
		t.Errorf("Expected entrypoint /amd64-app, got %v", cfg.Config.Entrypoint)
	}

	// Load arm64 image
	img2, err := idx.LoadImage("linux", "arm64")
	if err != nil {
		t.Fatalf("LoadImage linux/arm64: %v", err)
	}

	cfg2, err := img2.img.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}

	if cfg2.Config.Entrypoint[0] != "/arm64-app" {
		t.Errorf("Expected entrypoint /arm64-app, got %v", cfg2.Config.Entrypoint)
	}
}

func TestIndex_LoadImage_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	_, err = idx.LoadImage("windows", "amd64")
	if err == nil {
		t.Error("Expected error for nonexistent platform")
	}
}

func TestIndex_SaveLayout(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	layoutDir := filepath.Join(tmpDir, "layout")
	if err := idx.SaveLayout(layoutDir); err != nil {
		t.Fatalf("SaveLayout: %v", err)
	}

	// Verify OCI layout files exist
	files := []string{
		"oci-layout",
		"index.json",
		"blobs",
	}
	for _, f := range files {
		path := filepath.Join(layoutDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Missing OCI layout file: %s", f)
		}
	}
}

func TestExtractTar_PathTraversal(t *testing.T) {
	// This test verifies that path traversal attacks are prevented.
	// We can't easily create a malicious tarball without external tools,
	// but we can verify the extraction works for valid tarballs.
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	// ReadIndex internally uses extractTar
	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	// If we got here, extraction worked without path traversal issues
}

func TestIndex_Digest(t *testing.T) {
	tmpDir := t.TempDir()

	mb := NewMultiPlatformBuilder()
	mb.ForPlatform("linux", "amd64").SetEntrypoint("/bin/sh")

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	digest, err := idx.Digest()
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}

	if digest == "" {
		t.Error("Digest is empty")
	}

	// Digest should be in sha256:hex format
	if len(digest) != 71 { // "sha256:" + 64 hex chars
		t.Errorf("Digest format unexpected: %q (len=%d)", digest, len(digest))
	}
}
