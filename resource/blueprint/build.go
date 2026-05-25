package blueprint

import (
	"context"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource/affordance"
	"github.com/cruciblehq/crux/source"
)

// Resolves all services, compiles affordances, and writes the plan.
//
// The plan is written to plan.yaml in the output directory. The plan includes
// references to all services and their resolved grants, but does not include
// the service or runtime configs themselves. Those are expected to be pulled
// separately by the deployer when needed.
func Build(ctx context.Context, cfg *manifest.Blueprint, env string, src source.Source, output string) error {
	p, err := build(ctx, cfg, env, src, "localhost")
	if err != nil {
		return err
	}
	return manifest.WritePlanAt(p, output)
}

// Builds and writes the local deployment plan.
func BuildLocal(ctx context.Context, cfg *manifest.Blueprint, env string, src source.Source, output string) error {
	p, err := build(ctx, cfg, env, src, "localhost")
	if err != nil {
		return err
	}
	return manifest.WritePlanAt(p, output)
}

// Produces the deployment plan without writing it to disk.
//
// computeHost is set as the Host field of the local compute unit.
func build(ctx context.Context, cfg *manifest.Blueprint, envID string, src source.Source, computeHost string) (*manifest.Plan, error) {
	p := &manifest.Plan{
		Version: manifest.PlanVersion,
		Infrastructure: manifest.Infrastructure{
			Computes: map[string]manifest.Compute{"default": {Type: "local", Config: &manifest.ComputeLocal{Host: computeHost}}},
		},
		Services:     make(map[string]string),
		Containers:   make(map[string]manifest.Container),
		Environments: make(map[string]manifest.Environment),
	}

	p.Gateway = cfg.Gateway

	env, err := findEnvironment(cfg, envID)
	if err != nil {
		return nil, err
	}

	for _, service := range cfg.Services {
		target, ctr, dep, err := buildService(ctx, service, env, src)
		if err != nil {
			return nil, err
		}
		p.Services[service.ID] = target
		p.Containers[service.ID] = ctr
		p.Deployments = append(p.Deployments, dep)
	}

	p.Environments[env.ID] = *env
	deriveNetworks(p)
	return p, nil
}

// Resolves a single service reference into a target string, container spec, and
// deployment.
//
// Pulls the service manifest, validates the environment variables, and compiles
// affordances from the runtime and the service into a runtime spec.
func buildService(ctx context.Context, service manifest.Ref, env *manifest.Environment, src source.Source) (string, manifest.Container, manifest.Deployment, error) {
	ref, err := src.Parse(string(manifest.TypeService), service.Ref)
	if err != nil {
		return "", manifest.Container{}, manifest.Deployment{}, errService(service.ID, err)
	}

	serviceCfg, err := resolveService(ctx, service.ID, ref, src)
	if err != nil {
		return "", manifest.Container{}, manifest.Deployment{}, err
	}

	if err := validateEnvironment(serviceCfg.Schema, env); err != nil {
		return "", manifest.Container{}, manifest.Deployment{}, err
	}

	output := serviceCfg.OutputStage()
	if output == nil {
		return "", manifest.Container{}, manifest.Deployment{}, crex.Wrapf(ErrBuild, "service %s: no output stage", service.ID)
	}

	ctr, err := collectGrants(ctx, service.ID, output, src)
	if err != nil {
		return "", manifest.Container{}, manifest.Deployment{}, err
	}

	dep := manifest.Deployment{
		Service:     service.ID,
		Container:   service.ID,
		Environment: env.ID,
		Compute:     "default",
		Network:     "default",
	}
	return ref.String(), ctr, dep, nil
}

// Locates the selected environment in the blueprint.
//
// Returns an error if no environment with that ID is declared.
func findEnvironment(cfg *manifest.Blueprint, envID string) (*manifest.Environment, error) {
	for i := range cfg.Environments {
		if cfg.Environments[i].ID == envID {
			return &cfg.Environments[i], nil
		}
	}
	// When the blueprint declares no environments, no environment is required.
	if len(cfg.Environments) == 0 {
		return &manifest.Environment{ID: envID}, nil
	}
	return nil, crex.Wrapf(ErrBuild, "environment %q not found", envID)
}

// Checks that all required service schema parameters are present in the
// environment's variable map.
//
// Parameters with a default value are skipped.
func validateEnvironment(schema *manifest.Schema, env *manifest.Environment) error {
	if schema == nil {
		return nil
	}
	for _, p := range schema.Params {
		if p.Default != nil {
			continue
		}
		if _, ok := env.Variables[p.Name]; !ok {
			return crex.Wrapf(ErrBuild, "missing required variable %q", p.Name)
		}
	}
	return nil
}

// Compiles grants from the service's output stage and, when present,
// from its runtime into a compiled container.
//
// Runtime grants are processed before service-level grants.
func collectGrants(ctx context.Context, serviceID string, output *manifest.Stage, src source.Source) (manifest.Container, error) {
	var scopes []manifest.GrantScope
	if output.From != "" {
		rt, err := resolveRuntime(ctx, serviceID, output.From, src)
		if err != nil {
			return manifest.Container{}, err
		}
		if rtOutput := rt.OutputStage(); rtOutput != nil {
			scopes = append(scopes, rtOutput.Grants...)
		}
	}
	scopes = append(scopes, output.Grants...)

	b := affordance.NewBuilder()
	for _, scope := range scopes {
		for _, g := range scope.Grants {
			if err := b.Build(ctx, g, src); err != nil {
				return manifest.Container{}, crex.Wrapf(ErrBuild, "service %s: %w", serviceID, err)
			}
		}
	}
	return b.Spec().ToSpec(), nil
}

// Pulls a service resource and extracts its manifest config.
func resolveService(ctx context.Context, id string, ref *reference.Reference, src source.Source) (*manifest.Service, error) {
	result, err := src.Pull(ctx, ref)
	if err != nil {
		return nil, errService(id, err)
	}
	cfg, err := manifest.ReadAsAt[*manifest.Service](result.Extracted)
	if err != nil {
		return nil, errService(id, err)
	}
	return cfg, nil
}

// Pulls a runtime resource and extracts its manifest config.
func resolveRuntime(ctx context.Context, serviceID string, from string, src source.Source) (*manifest.Runtime, error) {
	ref, err := src.Parse(string(manifest.TypeRuntime), from)
	if err != nil {
		return nil, errServiceRuntime(serviceID, err)
	}
	result, err := src.Pull(ctx, ref)
	if err != nil {
		return nil, errServiceRuntime(serviceID, err)
	}
	cfg, err := manifest.ReadAsAt[*manifest.Runtime](result.Extracted)
	if err != nil {
		return nil, errServiceRuntime(serviceID, err)
	}
	return cfg, nil
}
