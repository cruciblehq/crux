package oci

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_SinglePlatform(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-builder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "app")
	if err := os.WriteFile(testFile, []byte("#!/bin/sh\necho hello"), 0755); err != nil {
		t.Fatal(err)
	}

	builder := NewBuilder("linux", "amd64")
	if err := builder.AddFile(testFile, "/app", 0755); err != nil {
		t.Fatalf("AddFile: %v", err)
	}
	builder.SetEntrypoint("/app")
	builder.SetEnv("MY_VAR", "my_value")
	builder.SetWorkdir("/")
	builder.SetLabel("version", "1.0.0")

	outputPath := filepath.Join(tmpDir, "image.tar")
	if err := builder.SaveTarball(outputPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Stat output: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output tarball is empty")
	}

	idx, err := ReadIndex(outputPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	manifest, err := idx.IndexManifest()
	if err != nil {
		t.Fatalf("IndexManifest: %v", err)
	}

	if manifest.SchemaVersion != 2 {
		t.Errorf("SchemaVersion = %d, want 2", manifest.SchemaVersion)
	}
	if len(manifest.Manifests) != 1 {
		t.Errorf("Manifests count = %d, want 1", len(manifest.Manifests))
	}
	if manifest.Manifests[0].Platform.OS != "linux" {
		t.Errorf("Platform OS = %q, want linux", manifest.Manifests[0].Platform.OS)
	}
	if manifest.Manifests[0].Platform.Architecture != "amd64" {
		t.Errorf("Platform Arch = %q, want amd64", manifest.Manifests[0].Platform.Architecture)
	}
}

func TestBuilder_AddBytes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-builder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	builder := NewBuilder("linux", "arm64")
	if err := builder.AddBytes([]byte(`{"key":"value"}`), "/config.json", 0644); err != nil {
		t.Fatalf("AddBytes: %v", err)
	}
	builder.SetCmd("cat", "/config.json")

	outputPath := filepath.Join(tmpDir, "image.tar")
	if err := builder.SaveTarball(outputPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(outputPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	manifest, err := idx.IndexManifest()
	if err != nil {
		t.Fatalf("IndexManifest: %v", err)
	}

	if manifest.Manifests[0].Platform.Architecture != "arm64" {
		t.Errorf("Platform Arch = %q, want arm64", manifest.Manifests[0].Platform.Architecture)
	}
}

func TestBuilder_AddDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-builder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.js"), []byte("console.log('hi')"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "lib"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "lib", "util.js"), []byte("module.exports = {}"), 0644); err != nil {
		t.Fatal(err)
	}

	builder := NewBuilder("linux", "amd64")
	if err := builder.AddLayer(srcDir, "/app"); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	builder.SetEntrypoint("node", "/app/main.js")

	outputPath := filepath.Join(tmpDir, "image.tar")
	if err := builder.SaveTarball(outputPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(outputPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()
}

func TestMultiPlatformBuilder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oci-builder-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	amd64Binary := filepath.Join(tmpDir, "app-amd64")
	arm64Binary := filepath.Join(tmpDir, "app-arm64")
	if err := os.WriteFile(amd64Binary, []byte("amd64 binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(arm64Binary, []byte("arm64 binary"), 0755); err != nil {
		t.Fatal(err)
	}

	mb := NewMultiPlatformBuilder()

	amd64 := mb.ForPlatform("linux", "amd64")
	if err := amd64.AddFile(amd64Binary, "/app", 0755); err != nil {
		t.Fatalf("AddFile amd64: %v", err)
	}
	amd64.SetEntrypoint("/app")
	amd64.SetLabel("version", "1.0.0")

	arm64 := mb.ForPlatform("linux", "arm64")
	if err := arm64.AddFile(arm64Binary, "/app", 0755); err != nil {
		t.Fatalf("AddFile arm64: %v", err)
	}
	arm64.SetEntrypoint("/app")
	arm64.SetLabel("version", "1.0.0")

	outputPath := filepath.Join(tmpDir, "image.tar")
	if err := mb.SaveTarball(outputPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(outputPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	manifest, err := idx.IndexManifest()
	if err != nil {
		t.Fatalf("IndexManifest: %v", err)
	}

	if len(manifest.Manifests) != 2 {
		t.Errorf("Manifests count = %d, want 2", len(manifest.Manifests))
	}

	platforms, err := idx.Platforms()
	if err != nil {
		t.Fatalf("Platforms: %v", err)
	}
	if !platforms["linux/amd64"] {
		t.Error("Missing linux/amd64 platform")
	}
	if !platforms["linux/arm64"] {
		t.Error("Missing linux/arm64 platform")
	}

	if err := idx.ValidateMultiPlatform(); err != nil {
		t.Errorf("ValidateMultiPlatform: %v", err)
	}
}
