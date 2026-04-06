package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/codec"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/reference"
)

// The canonical filename for plan output.
const planFile = "plan.yaml"

// [Builder] for Crucible blueprints.
//
// Building a blueprint resolves service references and their runtime
// affordances, producing a deployment [plan.Plan] as the build artifact.
// The plan is written to the output directory alongside the resolved
// manifest.
type BlueprintBuilder struct {
	source      Source
	environment string
}

// Returns a [BlueprintBuilder] configured with the given source.
//
// The environment selects which blueprint environment to include in the
// plan. Pass an empty string to omit environments.
func NewBlueprintBuilder(source Source, environment string) *BlueprintBuilder {
	return &BlueprintBuilder{source: source, environment: environment}
}

// Builds a Crucible blueprint resource based on the provided manifest.
//
// Each service reference is pulled from the registry and its runtime
// affordances (from the last build stage) are resolved into primitives.
// The resulting plan is written to the output directory as plan.yaml
// alongside the resolved manifest.
func (bb *BlueprintBuilder) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, err := manifestConfig[*manifest.Blueprint](&m)
	if err != nil {
		return nil, err
	}

	if _, err := reference.ParseIdentifier(m.Resource.Name, string(m.Resource.Type)); err != nil {
		return nil, crex.UserError("invalid resource name", "could not parse the resource identifier").
			Fallback("Check the resource name in crucible.yaml.").
			Cause(err).
			Err()
	}

	p, err := bb.compile(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := writePlan(p, output); err != nil {
		return nil, err
	}

	if err := WriteManifest(&m, output); err != nil {
		return nil, err
	}

	return &BuildResult{Output: output, Manifest: &m}, nil
}

// Verifies that the build directory contains the expected blueprint artifacts.
func (bb *BlueprintBuilder) Verify(buildDir string) error {
	return verify(buildDir, manifest.TypeBlueprint, planFile)
}

// Packages the blueprint's build output into a distributable archive.
func (bb *BlueprintBuilder) Pack(ctx context.Context, buildDir, output string) (*PackResult, error) {
	return pack(ctx, buildDir, output)
}

// Uploads a blueprint package archive to the Hub registry.
func (bb *BlueprintBuilder) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, bb.source.Registry, m, packagePath)
}

// Compiles a blueprint config into a deployment plan.
func (bb *BlueprintBuilder) compile(ctx context.Context, cfg *manifest.Blueprint) (*manifest.Plan, error) {
	p := &manifest.Plan{
		Version: manifest.PlanVersion,
		Compute: []manifest.Compute{{
			ID:       "default",
			Provider: "local",
		}},
	}

	p.Gateway = cfg.Gateway

	aff := NewAffordanceBuilder(bb.source)

	for _, svc := range cfg.Services {
		ref, afford, err := bb.resolveService(ctx, aff, svc)
		if err != nil {
			return nil, err
		}

		p.Services = append(p.Services, manifest.Ref{
			ID:     svc.ID,
			Target: ref,
		})

		ctr := manifest.Container{
			Service: svc.ID,
			Compute: "default",
		}
		if len(afford) > 0 {
			ctr.Grants = afford
		}
		if bb.environment != "" {
			ctr.Environment = bb.environment
		}
		p.Containers = append(p.Containers, ctr)
	}

	if bb.environment != "" {
		for _, env := range cfg.Environments {
			if env.ID == bb.environment {
				p.Environments = []manifest.Environment{env}
				break
			}
		}
	}

	return p, nil
}

// Pulls a service manifest and resolves its runtime affordances.
//
// Returns the fully qualified reference string and the resolved grants.
// Services with no stages or no runtime affordances return nil.
func (bb *BlueprintBuilder) resolveService(ctx context.Context, aff *AffordanceBuilder, svc manifest.Ref) (string, []manifest.Grant, error) {
	ref, err := bb.source.Parse(manifest.TypeService, svc.Target)
	if err != nil {
		return "", nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", svc.ID, err)
	}

	result, err := bb.source.Pull(ctx, ref)
	if err != nil {
		return "", nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", svc.ID, err)
	}

	m, err := ReadManifestIn(result.Dir)
	if err != nil {
		return "", nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", svc.ID, err)
	}

	cfg, err := manifestConfig[*manifest.Service](m)
	if err != nil {
		return "", nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", svc.ID, err)
	}

	stages := cfg.Stages
	if len(stages) == 0 {
		return ref.String(), nil, nil
	}

	last := stages[len(stages)-1]
	if len(last.Affordances) == 0 {
		return ref.String(), nil, nil
	}

	afford, err := aff.resolveRefs(ctx, last.Affordances)
	if err != nil {
		return "", nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", svc.ID, err)
	}

	return ref.String(), afford, nil
}

// Writes a plan to the given directory as plan.yaml.
func writePlan(p *manifest.Plan, dir string) error {
	data, err := codec.Encode(p, codec.YAML)
	if err != nil {
		return crex.Wrapf(ErrBlueprintBuild, "encode plan: %w", err)
	}
	path := filepath.Join(dir, planFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	return nil
}
