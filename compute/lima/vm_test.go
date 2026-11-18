//go:build darwin || linux

package lima

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/adrg/xdg"
)

func TestMachineArch(t *testing.T) {
	got := machineArch()
	if got != limaArchARM64 && got != limaArchAMD64 {
		t.Errorf("unexpected arch %q; want %q or %q", got, limaArchARM64, limaArchAMD64)
	}
}

func TestUploadImage_ExistingFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "machine*.qcow2")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()

	got, err := UploadImage(context.Background(), f.Name())
	if err != nil {
		t.Fatalf("UploadImage: %v", err)
	}
	if got != f.Name() {
		t.Errorf("got %q, want %q", got, f.Name())
	}
}

func TestUploadImage_MissingFile(t *testing.T) {
	_, err := UploadImage(context.Background(), "/nonexistent/path/machine.qcow2")
	if !errors.Is(err, ErrImageUpload) {
		t.Errorf("expected ErrImageUpload, got %v", err)
	}
}

func TestCachedMachineImagePath_MissingImage(t *testing.T) {
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	xdg.Reload()

	if _, err := cachedMachineImagePath(); !errors.Is(err, ErrMachineImageMissing) {
		t.Fatalf("cachedMachineImagePath() error = %v, want ErrMachineImageMissing", err)
	}
}

func TestFetchMachineImage_RequiresRegistryURL(t *testing.T) {
	if err := fetchMachineImage(context.Background(), nil); err == nil {
		t.Fatal("fetchMachineImage(nil) = nil, want non-nil error")
	}
}
