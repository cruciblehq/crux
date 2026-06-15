package resource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cruciblehq/crux/archive"
)

func TestPack(t *testing.T) {
	buildDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(buildDir, "manifest.yaml"), []byte("name: w"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(buildDir, "out.img"), []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(t.TempDir(), "dist", "pkg.tar.zst")
	res, err := Pack(t.Context(), buildDir, output)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	if res.Output != output {
		t.Fatalf("PackResult.Output = %q, want %q", res.Output, output)
	}

	// The archive round-trips back to the original files.
	extracted := filepath.Join(t.TempDir(), "out")
	if err := archive.Extract(output, extracted); err != nil {
		t.Fatalf("extract: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(extracted, "out.img")); err != nil || string(got) != "payload" {
		t.Fatalf("out.img = %q, %v; want %q", got, err, "payload")
	}
	if _, err := os.Stat(filepath.Join(extracted, "manifest.yaml")); err != nil {
		t.Fatalf("manifest.yaml missing after extract: %v", err)
	}
}

func TestEnsureOutputDir(t *testing.T) {
	root := t.TempDir()

	// A nested output path creates the parent directory.
	nested := filepath.Join(root, "a", "b", "pkg.tar.zst")
	if err := ensureOutputDir(nested); err != nil {
		t.Fatalf("ensureOutputDir nested: %v", err)
	}
	if info, err := os.Stat(filepath.Dir(nested)); err != nil || !info.IsDir() {
		t.Fatalf("parent dir not created: %v", err)
	}

	// A bare filename has no directory component and is a no-op.
	if err := ensureOutputDir("pkg.tar.zst"); err != nil {
		t.Fatalf("ensureOutputDir bare: %v", err)
	}
}
