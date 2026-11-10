package compute

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cruciblehq/crux/compute/local"
)

func TestStorageCreate(t *testing.T) {
	root := t.TempDir()
	id, err := newStorageLocal(root).Create(context.Background(), VolumeOptions{Mode: 0o750})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !strings.HasPrefix(id, local.VolumeIDPrefix) {
		t.Fatalf("Create() id = %q, want prefix %q", id, local.VolumeIDPrefix)
	}
	info, err := os.Stat(filepath.Join(root, local.VolumesDir, id))
	if err != nil {
		t.Fatalf("volume directory: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("volume %q is not a directory", id)
	}
}

func TestStorageCreateFails(t *testing.T) {
	root := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(root, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := newStorageLocal(root).Create(context.Background(), VolumeOptions{}); !errors.Is(err, ErrStorage) {
		t.Fatalf("Create() error = %v, want ErrStorage", err)
	}
}

func TestStorageListAndRemove(t *testing.T) {
	s := newStorageLocal(t.TempDir())
	id, err := s.Create(context.Background(), VolumeOptions{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	vols, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !slices.Equal(vols, []Volume{{ID: id}}) {
		t.Fatalf("List() = %v, want [%s]", vols, id)
	}

	if err := s.Remove(context.Background(), id); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	vols, err = s.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(vols) != 0 {
		t.Errorf("List() after Remove() = %v, want none", vols)
	}
}

func TestStorageListEmpty(t *testing.T) {
	vols, err := newStorageLocal(t.TempDir()).List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(vols) != 0 {
		t.Errorf("List() = %v, want none", vols)
	}
}

func TestStorageListFails(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, local.VolumesDir), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := newStorageLocal(root).List(context.Background()); !errors.Is(err, ErrStorage) {
		t.Fatalf("List() error = %v, want ErrStorage", err)
	}
}

func TestStorageAttachDetach(t *testing.T) {
	s := newStorageLocal(t.TempDir())
	if err := s.Attach(context.Background(), "vol-1", "crux"); err != nil {
		t.Errorf("Attach() error = %v", err)
	}
	if err := s.Detach(context.Background(), "vol-1"); err != nil {
		t.Errorf("Detach() error = %v", err)
	}
}
