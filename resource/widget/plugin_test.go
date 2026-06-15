package widget

import (
	"os"
	"path/filepath"
	"testing"

	es "github.com/evanw/esbuild/pkg/api"
)

func TestResolveEntryPoint(t *testing.T) {
	res, err := resolveModule(es.BuildOptions{}, es.OnResolveArgs{Path: "index.js", Kind: es.ResolveEntryPoint})
	if err != nil {
		t.Fatalf("resolveModule entry point: %v", err)
	}
	if res.Path != "index.js" || res.External {
		t.Fatalf("entry point result = %+v, want path index.js, not external", res)
	}
}

func TestResolveImportExternal(t *testing.T) {
	opts := es.BuildOptions{External: []string{"react", "@cruciblehq/ui"}}

	// Exact match is marked external.
	res, err := resolveImport(opts, es.OnResolveArgs{Path: "react"})
	if err != nil {
		t.Fatalf("resolveImport react: %v", err)
	}
	if !res.External || res.SideEffects != es.SideEffectsFalse {
		t.Fatalf("react result = %+v, want external with no side effects", res)
	}

	// Subpath of an external package is also external.
	res, err = resolveImport(opts, es.OnResolveArgs{Path: "@cruciblehq/ui/button"})
	if err != nil {
		t.Fatalf("resolveImport subpath: %v", err)
	}
	if !res.External {
		t.Fatalf("subpath result = %+v, want external", res)
	}
}

func TestResolveImportRelative(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "foo.js"), []byte("export default 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := resolveImport(es.BuildOptions{}, es.OnResolveArgs{Path: "./foo", ResolveDir: dir})
	if err != nil {
		t.Fatalf("resolveImport relative: %v", err)
	}
	if res.External {
		t.Fatalf("relative import marked external: %+v", res)
	}
	if res.Path != filepath.Join(dir, "foo.js") {
		t.Fatalf("resolved path = %q, want %q", res.Path, filepath.Join(dir, "foo.js"))
	}
}

func TestResolveImportFromDependencies(t *testing.T) {
	root := t.TempDir()
	deps := filepath.Join(root, dependenciesDirName)
	if err := os.MkdirAll(deps, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deps, "bar.js"), []byte("export default 2"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := es.BuildOptions{AbsWorkingDir: root}
	res, err := resolveImport(opts, es.OnResolveArgs{Path: "bar"})
	if err != nil {
		t.Fatalf("resolveImport dependency: %v", err)
	}
	if res.Path != filepath.Join(deps, "bar.js") {
		t.Fatalf("resolved path = %q, want %q", res.Path, filepath.Join(deps, "bar.js"))
	}
}
