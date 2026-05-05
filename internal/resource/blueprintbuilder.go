package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/internal/codec"
	"github.com/cruciblehq/crux/internal/crex"
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
// plan. It must match one of the environment IDs declared in the blueprint.
func NewBlueprintBuilder(source Source, environment string) *BlueprintBuilder {
	return &BlueprintBuilder{source: source, environment: environment}
}

// Builds a Crucible blueprint resource based on the provided manifest.
//
// References are pulled from the registry and their runtime affordances (from
// the last build stage) are resolved into primitives. The resulting plan is
// written to the output directory as plan.yaml alongside the resolved manifest.
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

	p, err := bb.build(ctx, cfg)
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

// Builds a blueprint config into a deployment plan.
func (bb *BlueprintBuilder) build(ctx context.Context, cfg *manifest.Blueprint) (*manifest.Plan, error) {
	p := &manifest.Plan{
		Version: manifest.PlanVersion,
		Compute: []manifest.Compute{{
			ID:       "default",
			Provider: "local",
		}},
	}

	p.Gateway = cfg.Gateway

	env, err := bb.findEnvironment(cfg)
	if err != nil {
		return nil, err
	}

	for _, service := range cfg.Services {
		svcRef, ctr, err := bb.buildService(ctx, service, env)
		if err != nil {
			return nil, err
		}
		p.Services = append(p.Services, svcRef)
		p.Containers = append(p.Containers, ctr)
	}

	p.Environments = []manifest.Environment{*env}

	return p, nil
}

// Resolves a single service reference into a plan service ref and container.
//
// Pulls the service manifest, extracts the output stage, collects affordances
// from the runtime and the service, and resolves them into flat grants. The
// service schema is validated against the environment.
func (bb *BlueprintBuilder) buildService(ctx context.Context, service manifest.Ref, env *manifest.Environment) (manifest.Ref, manifest.Container, error) {
	ref, err := bb.source.Parse(manifest.TypeService, service.Target)
	if err != nil {
		return manifest.Ref{}, manifest.Container{}, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", service.ID, err)
	}

	serviceCfg, err := bb.resolveService(ctx, service.ID, ref)
	if err != nil {
		return manifest.Ref{}, manifest.Container{}, err
	}

	if err := validateEnvironment(service.ID, serviceCfg.Schema, env); err != nil {
		return manifest.Ref{}, manifest.Container{}, err
	}

	output := serviceCfg.OutputStage()
	if output == nil {
		return manifest.Ref{}, manifest.Container{}, crex.Wrapf(ErrBlueprintBuild, "service %s: no output stage", service.ID)
	}

	affordances, err := bb.collectAffordances(ctx, service.ID, output)
	if err != nil {
		return manifest.Ref{}, manifest.Container{}, err
	}

	grants, err := bb.resolveAffordances(ctx, service.ID, affordances)
	if err != nil {
		return manifest.Ref{}, manifest.Container{}, err
	}

	svcRef := manifest.Ref{
		ID:     service.ID,
		Target: ref.String(),
	}

	ctr := manifest.Container{
		Service:     service.ID,
		Compute:     "default",
		Grants:      grants,
		Environment: env.ID,
	}

	return svcRef, ctr, nil
}

// Finds the selected environment in the blueprint.
//
// Returns an error if the environment is not declared in the blueprint.
func (bb *BlueprintBuilder) findEnvironment(cfg *manifest.Blueprint) (*manifest.Environment, error) {
	for i := range cfg.Environments {
		if cfg.Environments[i].ID == bb.environment {
			return &cfg.Environments[i], nil
		}
	}
	return nil, crex.Wrapf(ErrBlueprintBuild, "environment %q not found", bb.environment)
}

// Validates that an environment provides all required variables for a service.
//
// Required parameters (those without a default value) must have a matching
// key in the environment's variable map. Optional parameters are skipped.
func validateEnvironment(serviceID string, schema manifest.Schema, env *manifest.Environment) error {
	for _, p := range schema.Params {
		if p.Default != nil {
			continue
		}
		if _, ok := env.Variables[p.Name]; !ok {
			return crex.Wrapf(ErrBlueprintBuild, "service %s: environment %q: missing required variable %q", serviceID, env.ID, p.Name)
		}
	}
	return nil
}

// Collects affordances from a service's output stage and its runtime.
//
// When the stage has a From ref, the runtime is pulled and its output stage
// affordances are prepended. The service's own affordances are always appended.
func (bb *BlueprintBuilder) collectAffordances(ctx context.Context, serviceID string, output *manifest.Stage) ([]manifest.Ref, error) {
	var affordances []manifest.Ref
	if output.From != nil {
		runtime, err := bb.resolveRuntime(ctx, serviceID, output.From)
		if err != nil {
			return nil, err
		}
		if runtimeOutput := runtime.OutputStage(); runtimeOutput != nil {
			affordances = append(affordances, runtimeOutput.Affordances...)
		}
	}
	return append(affordances, output.Affordances...), nil
}

// Pulls a service manifest and returns its config.
func (bb *BlueprintBuilder) resolveService(ctx context.Context, id string, ref *reference.Reference) (*manifest.Service, error) {
	result, err := bb.source.Pull(ctx, ref)
	if err != nil {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", id, err)
	}

	m, err := ReadManifestIn(result.Dir)
	if err != nil {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", id, err)
	}

	cfg, ok := m.Config.(*manifest.Service)
	if !ok {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: not a service manifest", id)
	}

	return cfg, nil
}

// Pulls a runtime manifest and returns its config.
func (bb *BlueprintBuilder) resolveRuntime(ctx context.Context, serviceID string, from *manifest.Ref) (*manifest.Runtime, error) {
	ref, err := bb.source.Parse(manifest.TypeRuntime, from.Target)
	if err != nil {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: runtime: %w", serviceID, err)
	}

	result, err := bb.source.Pull(ctx, ref)
	if err != nil {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: runtime: %w", serviceID, err)
	}

	m, err := ReadManifestIn(result.Dir)
	if err != nil {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: runtime: %w", serviceID, err)
	}

	cfg, ok := m.Config.(*manifest.Runtime)
	if !ok {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %s is not a runtime manifest", serviceID, from.Target)
	}

	return cfg, nil
}

// Resolves affordance references into a flat grant list.
//
// Each ref is pulled from the registry and its grant scopes are collected.
// The combined scopes are then resolved through the affordance builder,
// which expands nested references and builds domain grants via subsystems.
func (bb *BlueprintBuilder) resolveAffordances(ctx context.Context, serviceID string, refs []manifest.Ref) ([]manifest.Grant, error) {
	if len(refs) == 0 {
		return nil, nil
	}

	var scopes []manifest.GrantScope
	for _, ref := range refs {
		aff, _, err := pullAffordance(ctx, bb.source, ref.Target)
		if err != nil {
			return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: affordance %s: %w", serviceID, ref.Target, err)
		}
		scopes = append(scopes, aff.Scopes...)
	}

	ab := NewAffordanceBuilder(bb.source)
	resolved, err := ab.resolve(ctx, scopes)
	if err != nil {
		return nil, crex.Wrapf(ErrBlueprintBuild, "service %s: %w", serviceID, err)
	}

	var grants []manifest.Grant
	for _, scope := range resolved {
		grants = append(grants, scope.Grants...)
	}
	return grants, nil
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
