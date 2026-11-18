package compute

import (
	"context"
	"errors"
	"testing"

	"github.com/adrg/xdg"
)

func TestMachineArch(t *testing.T) {
	got := machineArch()
	if got != imageArchARM64 && got != imageArchAMD64 {
		t.Errorf("unexpected arch %q; want %q or %q", got, imageArchARM64, imageArchAMD64)
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
