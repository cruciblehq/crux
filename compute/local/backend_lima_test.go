//go:build darwin || linux

package local

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestUploadImage_ExistingFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "machine*.qcow2")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()

	got, err := uploadImage(context.Background(), f.Name())
	if err != nil {
		t.Fatalf("uploadImage: %v", err)
	}
	if got != f.Name() {
		t.Errorf("got %q, want %q", got, f.Name())
	}
}

func TestUploadImage_MissingFile(t *testing.T) {
	_, err := uploadImage(context.Background(), "/nonexistent/path/machine.qcow2")
	if !errors.Is(err, ErrImageUpload) {
		t.Errorf("expected ErrImageUpload, got %v", err)
	}
}
