package oci

import (
	"path/filepath"
	"testing"
)

func TestImage_Digest(t *testing.T) {
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

	img, err := idx.LoadImage("linux", "amd64")
	if err != nil {
		t.Fatalf("LoadImage: %v", err)
	}

	digest, err := img.Digest()
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}

	if digest == "" {
		t.Error("Digest is empty")
	}

	if len(digest) != 71 {
		t.Errorf("Digest format unexpected: %q (len=%d)", digest, len(digest))
	}
}

func TestImage_Layers(t *testing.T) {
	tmpDir := t.TempDir()

	builder := NewBuilder("linux", "amd64")
	if err := builder.AddBytes([]byte("file1"), "/file1.txt", 0644); err != nil {
		t.Fatal(err)
	}
	if err := builder.AddBytes([]byte("file2"), "/file2.txt", 0644); err != nil {
		t.Fatal(err)
	}

	tarPath := filepath.Join(tmpDir, "image.tar")
	if err := builder.SaveTarball(tarPath); err != nil {
		t.Fatalf("SaveTarball: %v", err)
	}

	idx, err := ReadIndex(tarPath)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	defer idx.Close()

	img, err := idx.LoadImage("linux", "amd64")
	if err != nil {
		t.Fatalf("LoadImage: %v", err)
	}

	layers, err := img.Layers()
	if err != nil {
		t.Fatalf("Layers: %v", err)
	}

	if len(layers) != 2 {
		t.Fatalf("Expected 2 layers, got %d", len(layers))
	}

	for i, l := range layers {
		if l.Digest == "" {
			t.Errorf("Layer %d: digest is empty", i)
		}
		if l.Size <= 0 {
			t.Errorf("Layer %d: size is %d, expected > 0", i, l.Size)
		}
		if l.MediaType == "" {
			t.Errorf("Layer %d: media type is empty", i)
		}
	}
}
