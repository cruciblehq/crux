package compute

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/cruciblehq/crux/compute/lima"
	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/utils-go/file"
)

func useTempDir(t *testing.T) string {
	t.Helper()
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	xdg.Reload()
	return dirs.Temp()
}

func TestDefaultLocalMachineID(t *testing.T) {
	if got := DefaultLocalMachineID(); got != localMachineName {
		t.Fatalf("DefaultLocalMachineID() = %q, want %q", got, localMachineName)
	}
}

func TestLimaStateToState(t *testing.T) {
	if got, want := limaStateToState(lima.StateRunning), StateRunning; got != want {
		t.Fatalf("limaStateToState() = %v, want %v", got, want)
	}
	if got, want := limaStateToState(lima.StateStopped), StateStopped; got != want {
		t.Fatalf("limaStateToState() = %v, want %v", got, want)
	}
	if got, want := limaStateToState(lima.StateNotProvisioned), StateNotProvisioned; got != want {
		t.Fatalf("limaStateToState() = %v, want %v", got, want)
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
	if info.Mode().Perm() != file.OwnerDirMode {
		t.Errorf("temp dir mode = %v, want %v", info.Mode().Perm(), file.OwnerDirMode)
	}
}

func TestUploadImageFromReader(t *testing.T) {
	dir := useTempDir(t)
	var p localProvider

	id, err := p.UploadImage(context.Background(), bytes.NewReader([]byte("disk")))
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

	if err := p.RemoveImage(context.Background(), id); err != nil {
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

	var p localProvider
	id, err := p.UploadImage(context.Background(), f)
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}
	if id != path {
		t.Errorf("UploadImage() id = %q, want %q", id, path)
	}

	if err := p.RemoveImage(context.Background(), id); err != nil {
		t.Fatalf("RemoveImage() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("image outside the staging directory was removed; stat error = %v", err)
	}
}
