package resource

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cruciblehq/crux/manifest"
)

func writeWidgetManifest(t *testing.T, dir string) {
	t.Helper()
	m := &manifest.Manifest{
		Resource: manifest.Resource{Type: manifest.TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &manifest.Widget{Main: "index.js"},
	}
	if err := manifest.WriteAt(m, dir); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func TestVerifyUnsupportedType(t *testing.T) {
	if err := Verify("/build", manifest.ResourceType("bogus")); !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("Verify(bogus) = %v, want ErrUnsupportedType", err)
	}
}

func TestVerify(t *testing.T) {
	dir := t.TempDir()
	writeWidgetManifest(t, dir)

	if err := verify(dir, manifest.TypeWidget, ""); err != nil {
		t.Fatalf("verify without artifact: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "out.img"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verify(dir, manifest.TypeWidget, "out.img"); err != nil {
		t.Fatalf("verify with artifact: %v", err)
	}

	if err := verify(dir, manifest.TypeWidget, "missing.img"); !errors.Is(err, ErrBuildOutputNotFound) {
		t.Fatalf("verify missing artifact = %v, want ErrBuildOutputNotFound", err)
	}
}

func TestVerifyBuildDir(t *testing.T) {
	dir := t.TempDir()
	writeWidgetManifest(t, dir)

	if _, err := verifyBuildDir(dir, manifest.TypeWidget); err != nil {
		t.Fatalf("verifyBuildDir matching type: %v", err)
	}

	if _, err := verifyBuildDir(dir, manifest.TypeService); !errors.Is(err, ErrResourceTypeMismatch) {
		t.Fatalf("verifyBuildDir mismatch = %v, want ErrResourceTypeMismatch", err)
	}

	if _, err := verifyBuildDir(t.TempDir(), manifest.TypeWidget); !errors.Is(err, ErrManifestNotFound) {
		t.Fatalf("verifyBuildDir missing = %v, want ErrManifestNotFound", err)
	}
}
