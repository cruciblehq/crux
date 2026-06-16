package blueprint

import (
	"errors"
	"testing"

	"github.com/cruciblehq/crux/affordance/kernel"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/registry"
)

func TestNewBuilder(t *testing.T) {
	src, err := registry.NewSource("http://reg", "ns")
	if err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(src, "prod")
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if b.src != src {
		t.Errorf("src = %+v, want %+v", b.src, src)
	}
	if b.env != "prod" {
		t.Errorf("env = %q, want prod", b.env)
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

	// With no compute units, services map to the empty assignment.
	empty := binPack(results, map[string]manifest.Compute{})
	if empty["a"] != "" || empty["b"] != "" {
		t.Fatalf("binPack empty computes = %v, want empty assignments", empty)
	}
}

func TestDeriveComputeKernel(t *testing.T) {
	results := []serviceResult{
		{serviceID: "a", kernel: kernel.Spec{Features: []string{"NETFILTER"}}},
		{serviceID: "b", kernel: kernel.Spec{Features: []string{"FUSE_FS"}}},
	}
	assignments := map[string]string{"a": "c1", "b": "c2"}

	spec := deriveComputeKernel("c1", assignments, results)

	// Only the service assigned to c1 contributes its kernel requirements.
	if len(spec.Features) != 1 || spec.Features[0] != "NETFILTER" {
		t.Fatalf("c1 features = %v, want [NETFILTER]", spec.Features)
	}
}
