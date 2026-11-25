package blueprint

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/registry"
)

const serviceManifest = `version: 0
resource:
  type: service
  name: ns/api
  version: 1.0.0
entrypoint: ["/bin/api"]
stages:
  - name: service
    from: ns/base 1.0.0
    steps:
      - run: "true"
`

func writeServiceManifest(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifest.ManifestFile), []byte(serviceManifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Points at a registry that does not exist, so any registry call fails.
func unreachableSource(t *testing.T) registry.Source {
	t.Helper()
	src, err := registry.NewSource("http://127.0.0.1:1", "ns", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func TestResolvePlanPathService(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "api")
	writeServiceManifest(t, dir)
	cfg := &manifest.Blueprint{Services: []manifest.Ref{{ID: "api", Path: dir}}}

	p, err := resolvePlan(context.Background(), cfg, "dev", unreachableSource(t), "/elsewhere", defaultCompute())
	if err != nil {
		t.Fatalf("resolvePlan: %v", err)
	}
	if p.Services["api"] != dir {
		t.Errorf("Services[api] = %q, want %q", p.Services["api"], dir)
	}
	if len(p.Deployments) != 1 || p.Deployments[0].Service != "api" {
		t.Errorf("Deployments = %+v, want one deployment of api", p.Deployments)
	}
}

func TestResolvePlanRelativePath(t *testing.T) {
	base := t.TempDir()
	writeServiceManifest(t, filepath.Join(base, "services", "api"))
	cfg := &manifest.Blueprint{Services: []manifest.Ref{{ID: "api", Path: "services/api"}}}

	p, err := resolvePlan(context.Background(), cfg, "dev", unreachableSource(t), base, defaultCompute())
	if err != nil {
		t.Fatalf("resolvePlan: %v", err)
	}
	want := filepath.Join(base, "services", "api")
	if p.Services["api"] != want {
		t.Errorf("Services[api] = %q, want %q", p.Services["api"], want)
	}
}

func TestResolvePlanPathWithoutManifest(t *testing.T) {
	dir := t.TempDir()
	cfg := &manifest.Blueprint{Services: []manifest.Ref{{ID: "api", Path: dir}}}

	_, err := resolvePlan(context.Background(), cfg, "dev", unreachableSource(t), "/elsewhere", defaultCompute())
	if !errors.Is(err, ErrBuildPlan) {
		t.Fatalf("resolvePlan = %v, want ErrBuildPlan", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("resolvePlan = %v, want os.ErrNotExist", err)
	}
}

func TestNewBuilder(t *testing.T) {
	src, err := registry.NewSource("http://reg", "ns", "/cache")
	if err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(src, "/work", "prod", manifest.Compute{})
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if b.src != src {
		t.Errorf("src = %+v, want %+v", b.src, src)
	}
	if b.dir != "/work" {
		t.Errorf("dir = %q, want /work", b.dir)
	}
	if b.env != "prod" {
		t.Errorf("env = %q, want prod", b.env)
	}
	if b.compute.Type != manifest.ComputeTypeLocal {
		t.Errorf("compute.Type = %q, want %q", b.compute.Type, manifest.ComputeTypeLocal)
	}
	localCfg, ok := b.compute.Config.(*manifest.ComputeLocal)
	if !ok {
		t.Fatalf("compute.Config = %T, want *manifest.ComputeLocal", b.compute.Config)
	}
	if localCfg.Host != localComputeHost {
		t.Errorf("compute host = %q, want %q", localCfg.Host, localComputeHost)
	}

	custom := NewBuilder(src, "/work", "prod", manifest.Compute{Type: manifest.ComputeTypeLocal, Config: &manifest.ComputeLocal{Host: "example.com"}})
	customCfg, ok := custom.compute.Config.(*manifest.ComputeLocal)
	if !ok {
		t.Fatalf("custom compute.Config = %T, want *manifest.ComputeLocal", custom.compute.Config)
	}
	if customCfg.Host != "example.com" {
		t.Errorf("custom host = %q, want example.com", customCfg.Host)
	}
}

func TestFindEnvironment(t *testing.T) {
	cfg := &manifest.Blueprint{
		Environments: []manifest.Environment{
			{ID: "dev"},
			{ID: "prod"},
		},
	}

	env, err := findEnvironment(cfg, "prod")
	if err != nil {
		t.Fatalf("findEnvironment prod: %v", err)
	}
	if env.ID != "prod" {
		t.Fatalf("env.ID = %q, want prod", env.ID)
	}

	if _, err := findEnvironment(cfg, "missing"); !errors.Is(err, ErrBuildPlan) {
		t.Fatalf("findEnvironment missing = %v, want ErrBuildPlan", err)
	}

	// A blueprint with no environments yields a synthetic empty one.
	empty := &manifest.Blueprint{}
	env, err = findEnvironment(empty, "any")
	if err != nil {
		t.Fatalf("findEnvironment no-envs: %v", err)
	}
	if env.ID != "any" {
		t.Fatalf("synthetic env.ID = %q, want any", env.ID)
	}
}

func TestValidateEnvironment(t *testing.T) {
	// A nil schema requires nothing.
	if err := validateEnvironment(nil, &manifest.Environment{}); err != nil {
		t.Fatalf("nil schema: %v", err)
	}

	schema := &manifest.Schema{
		Params: []manifest.Param{
			{Name: "required"},
			{Name: "optional", Default: "x"},
		},
	}

	// Missing required variable fails.
	if err := validateEnvironment(schema, &manifest.Environment{}); !errors.Is(err, ErrBuildPlan) {
		t.Fatalf("missing required = %v, want ErrBuildPlan", err)
	}

	// Required variable present, optional omitted: valid.
	env := &manifest.Environment{Variables: map[string]string{"required": "v"}}
	if err := validateEnvironment(schema, env); err != nil {
		t.Fatalf("required present: %v", err)
	}
}

func TestBinPack(t *testing.T) {
	results := []serviceResult{
		{serviceID: "a"},
		{serviceID: "b"},
	}
	computes := map[string]manifest.Compute{"only": {}}

	got := binPack(results, computes)
	if got["a"] != "only" || got["b"] != "only" {
		t.Fatalf("binPack = %v, want both on \"only\"", got)
	}

	// With no compute entries, services map to the empty assignment.
	empty := binPack(results, map[string]manifest.Compute{})
	if empty["a"] != "" || empty["b"] != "" {
		t.Fatalf("binPack empty computes = %v, want empty assignments", empty)
	}
}
