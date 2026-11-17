package machines

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cruciblehq/utils-go/file"
)

func TestNew(t *testing.T) {
	r := New()
	if err := r.ID.Validate(); err != nil {
		t.Errorf("ID.Validate() = %v, want nil", err)
	}
	if r.Enrolled() {
		t.Error("Enrolled() = true, want false")
	}
}

func TestSaveLoad(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "machines")
	r := New()

	if err := Save(dir, "crux", r); err != nil {
		t.Fatalf("Save() = %v, want nil", err)
	}
	got, err := Load(dir, "crux")
	if err != nil {
		t.Fatalf("Load() = %v, want nil", err)
	}
	if got.ID != r.ID || string(got.Token) != string(r.Token) {
		t.Errorf("Load() = %+v, want %+v", got, r)
	}
	info, err := os.Stat(filepath.Join(dir, "crux.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != file.OwnerFileMode {
		t.Errorf("mode = %v, want %v", info.Mode().Perm(), file.OwnerFileMode)
	}
}

func TestSaveEnrolled(t *testing.T) {
	dir := t.TempDir()
	r := New()
	r.Token = nil

	if err := Save(dir, "crux", r); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir, "crux")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enrolled() {
		t.Error("Enrolled() = false, want true")
	}
}

func TestLoadNotFound(t *testing.T) {
	if _, err := Load(t.TempDir(), "crux"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() = %v, want %v", err, ErrNotFound)
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "crux.json"), []byte(`{"id":"bad"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir, "crux"); !errors.Is(err, ErrInvalidRecord) {
		t.Errorf("Load() = %v, want %v", err, ErrInvalidRecord)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, "crux", New()); err != nil {
		t.Fatal(err)
	}

	if err := Remove(dir, "crux"); err != nil {
		t.Fatalf("Remove() = %v, want nil", err)
	}
	if _, err := Load(dir, "crux"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() after Remove = %v, want %v", err, ErrNotFound)
	}
	if err := Remove(dir, "crux"); err != nil {
		t.Errorf("second Remove() = %v, want nil", err)
	}
}
