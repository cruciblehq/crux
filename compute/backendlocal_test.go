package compute

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/cruciblehq/crux/compute/local"
	"github.com/cruciblehq/crux/internal/dirs"
)

func useTempDir(t *testing.T) string {
	t.Helper()
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	xdg.Reload()
	return dirs.Temp()
}

func TestDefaultLocalInstanceID(t *testing.T) {
	if got := DefaultLocalInstanceID(); got != localInstanceName {
		t.Fatalf("DefaultLocalInstanceID() = %q, want %q", got, localInstanceName)
	}
}

func TestLocalStateToState(t *testing.T) {
	if got, want := localStateToState(local.StateRunning), StateRunning; got != want {
		t.Fatalf("localStateToState() = %v, want %v", got, want)
	}
	if got, want := localStateToState(local.StateStopped), StateStopped; got != want {
		t.Fatalf("localStateToState() = %v, want %v", got, want)
	}
	if got, want := localStateToState(local.StateNotProvisioned), StateNotProvisioned; got != want {
		t.Fatalf("localStateToState() = %v, want %v", got, want)
	}
}

func TestNewBackendLocal(t *testing.T) {
	b := newBackendLocal()
	if b.Compute() == nil {
		t.Error("Compute() = nil, want a compute manager")
	}
	if b.Storage() == nil {
		t.Error("Storage() = nil, want a storage manager")
	}
	if b.Network() == nil {
		t.Error("Network() = nil, want a network manager")
	}
}

func TestCreateTemp(t *testing.T) {
	dir := useTempDir(t)
	f, err := createTemp("x-*")
	if err != nil {
		t.Fatalf("createTemp() error = %v", err)
	}
	defer f.Close()
	if filepath.Dir(f.Name()) != dir {
		t.Errorf("createTemp() file = %q, want it in %q", f.Name(), dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != tempDirMode {
		t.Errorf("temp dir mode = %v, want %v", info.Mode().Perm(), tempDirMode)
	}
}

func TestUploadImageFromReader(t *testing.T) {
	dir := useTempDir(t)
	shim := &computeLocalShim{local: local.NewBackend()}

	id, err := shim.UploadImage(context.Background(), bytes.NewReader([]byte("disk")))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}
	if !strings.HasPrefix(id, dir) {
		t.Errorf("UploadImage() id = %q, want it in %q", id, dir)
	}
	data, err := os.ReadFile(id)
	if err != nil {
		t.Fatalf("read staged image: %v", err)
	}
	if string(data) != "disk" {
		t.Errorf("staged image = %q, want %q", data, "disk")
	}

	if err := shim.RemoveImage(context.Background(), id); err != nil {
		t.Fatalf("RemoveImage() error = %v", err)
	}
	if _, err := os.Stat(id); !os.IsNotExist(err) {
		t.Errorf("staged image still exists after RemoveImage(); stat error = %v", err)
	}
}

func TestUploadImageFromFile(t *testing.T) {
	useTempDir(t)
	path := filepath.Join(t.TempDir(), "machine.qcow2")
	if err := os.WriteFile(path, []byte("disk"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	shim := &computeLocalShim{local: local.NewBackend()}
	id, err := shim.UploadImage(context.Background(), f)
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}
	if id != path {
		t.Errorf("UploadImage() id = %q, want %q", id, path)
	}

	if err := shim.RemoveImage(context.Background(), id); err != nil {
		t.Fatalf("RemoveImage() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("image outside the staging directory was removed; stat error = %v", err)
	}
}
