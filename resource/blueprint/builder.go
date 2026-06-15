package blueprint

import (
	"context"

	"github.com/cruciblehq/crux/affordance/kernel"
	"github.com/cruciblehq/crux/affordance/provision"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/registry"
	"github.com/cruciblehq/crux/resource/affordance"
)

// Builds a Crucible blueprint resource from its configuration.
type Builder struct {
	src registry.Source // Registry access for resolving service references and affordances.
	env string          // Environment selector for the blueprint.
}

// Returns a new Builder.
//
// source provides registry access for resolving service references and their
// affordances. env selects the blueprint environment to resolve.
func NewBuilder(src registry.Source, env string) *Builder {
	return &Builder{src: src, env: env}
}

// Builds the blueprint described by cfg.
//
// Writes plan.yaml to the output directory. The plan includes references to
// all services and their resolved grants, but does not include the service or
// runtime configs themselves. Those are expected to be pulled separately by the
// deployer when needed.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Blueprint, output string) error {
	p, err := resolvePlan(ctx, cfg, b.env, b.src, "localhost")
	if err != nil {
		return err
	}
	return manifest.WritePlanAt(p, output)
}

// Produces the deployment plan for a blueprint without writing it to disk.
//
// The selected environment is located by envID, then every service is resolved
// against src: its manifest is pulled, its environment variables are validated,
// and its affordance grants are compiled into a container spec with kernel and
// provision requirements. Resolved services are bin-packed onto compute units,
// and each compute's security model is derived from the union of the kernel
// requirements of the services assigned to it. A deployment is emitted per
// service binding it to its compute and environment, and the cloud network
// perimeter is derived from the per-container network specs. computeHost is set
// as the Host of the single local compute unit. Returns the assembled plan, or
// an error from environment lookup or any service resolution.
func resolvePlan(ctx context.Context, cfg *manifest.Blueprint, envID string, src registry.Source, computeHost string) (*manifest.Plan, error) {
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

	// Resolve all services and accumulate their affordances.
	results := make([]serviceResult, 0, len(cfg.Services))
	for _, service := range cfg.Services {
		r, err := planService(ctx, service, env, src)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
		p.Services[r.serviceID] = r.target
		p.Containers[r.serviceID] = r.container
	}

	// Assign services to compute units before deriving per-compute kernel requirements.
	assignments := binPack(results, p.Infrastructure.Computes)

	// Derive the kernel requirements per compute unit from the union of assigned services.
	for computeID := range p.Infrastructure.Computes {
		model := deriveComputeSecurityModel(computeID, assignments, results)
		compute := p.Infrastructure.Computes[computeID]
		compute.Security = &model
		p.Infrastructure.Computes[computeID] = compute
	}

	// Build deployments using the bin-packed assignments.
	for _, r := range results {
		computeID := assignments[r.serviceID]
		p.Deployments = append(p.Deployments, manifest.Deployment{
			Service:     r.serviceID,
			Container:   r.serviceID,
			Environment: env.ID,
			Compute:     computeID,
			Network:     computeID,
		})
	}

	p.Environments[env.ID] = *env
	deriveNetworks(p)
	return p, nil
}

// Resolved affordance data for a single service.
type serviceResult struct {
	serviceID string             // manifest service identifier
	target    string             // resolved reference string for the service image
	container manifest.Container // compiled container spec from grants
	kernel    kernel.Spec        // kernel requirements accumulated from grants
	cpuMillis uint64             // provisioned CPU in milli-cores
	memBytes  uint64             // provisioned memory in bytes
	diskBytes uint64             // provisioned disk in bytes
}

// Resolves a single service reference into a serviceResult.
//
// Pulls the service manifest, validates the environment variables, and compiles
// affordances from the runtime and the service into a runtime spec.
func planService(ctx context.Context, service manifest.Ref, env *manifest.Environment, src registry.Source) (serviceResult, error) {
	ref, err := src.Parse(string(manifest.TypeService), service.Ref)
	if err != nil {
		return serviceResult{}, errService(service.ID, err)
	}

	serviceCfg, err := resolveService(ctx, service.ID, ref, src)
	if err != nil {
		return serviceResult{}, err
	}

	if err := validateEnvironment(serviceCfg.Schema, env); err != nil {
		return serviceResult{}, err
	}

	output := serviceCfg.OutputStage()
	if output == nil {
		return serviceResult{}, crex.Wrapf(ErrBuildPlan, "service %s: no output stage", service.ID)
	}

	ctr, kspec, prov, err := collectGrants(ctx, service.ID, output, src)
	if err != nil {
		return serviceResult{}, err
	}

	return serviceResult{
		serviceID: service.ID,
		target:    ref.String(),
		container: ctr,
		kernel:    kspec,
		cpuMillis: prov.CPU,
		memBytes:  prov.Memory,
		diskBytes: prov.Disk,
	}, nil
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
	return nil, crex.Wrapf(ErrBuildPlan, "environment %q not found", envID)
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
			return crex.Wrapf(ErrBuildPlan, "missing required variable %q", p.Name)
		}
	}
	return nil
}

// Compiles grants from the service's output stage and, when present,
// from its runtime into a compiled container, kernel spec, and provision spec.
//
// Runtime grants are processed before service-level grants.
func collectGrants(ctx context.Context, serviceID string, output *manifest.Stage, src registry.Source) (manifest.Container, kernel.Spec, *provision.Spec, error) {
	var scopes []manifest.GrantScope
	if output.From != "" {
		rt, err := resolveRuntime(ctx, serviceID, output.From, src)
		if err != nil {
			return manifest.Container{}, kernel.Spec{}, nil, err
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
				return manifest.Container{}, kernel.Spec{}, nil, crex.Wrapf(ErrBuildPlan, "service %s: %w", serviceID, err)
			}
		}
	}

	var kspec kernel.Spec
	if b.Spec().Kernel != nil {
		kspec = *b.Kernel()
	}
	return assembleContainer(b), kspec, b.Provision(), nil
}

// Assembles a container spec from the sections the builder produced.
//
// Each section is copied only when present; absent sections leave the
// container's corresponding zero value in place.
func assembleContainer(b *affordance.Builder) manifest.Container {
	s := b.Spec()
	container := manifest.Container{}
	if s.OCI != nil {
		container.OCI = *s.OCI
	}
	if s.Fcap != nil {
		container.Fcap = *s.Fcap
	}
	if s.MAC != nil {
		container.MAC = *s.MAC
	}
	if s.Net != nil {
		container.Network = *s.Net
	}
	if s.Volume != nil {
		container.Volumes = s.Volume.Mounts
	}
	return container
}

// Assigns services to compute units using first-fit bin-packing.
//
// All services are assigned to the first compute unit in iteration order.
// This is a simple placeholder that satisfies the interface; a more
// sophisticated packer can be substituted here without changing callers.
func binPack(results []serviceResult, computes map[string]manifest.Compute) map[string]string {
	var computeID string
	for id := range computes {
		computeID = id
		break
	}
	assignments := make(map[string]string, len(results))
	for _, r := range results {
		assignments[r.serviceID] = computeID
	}
	return assignments
}

// Derives the kernel-level ComputeSecurityModel for a single compute unit.
//
// Unions the kernel specs of all services assigned to computeID. Entries are
// deduplicated.
func deriveComputeSecurityModel(computeID string, assignments map[string]string, results []serviceResult) manifest.ComputeSecurityModel {
	var model manifest.ComputeSecurityModel
	for _, r := range results {
		if assignments[r.serviceID] != computeID {
			continue
		}
		model.Kernel.Merge(r.kernel)
	}
	return model
}

// Pulls a service resource and extracts its manifest config.
func resolveService(ctx context.Context, id string, ref *reference.Reference, src registry.Source) (*manifest.Service, error) {
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
func resolveRuntime(ctx context.Context, serviceID string, from string, src registry.Source) (*manifest.Runtime, error) {
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
