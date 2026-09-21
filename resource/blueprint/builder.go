package blueprint

import (
	"context"

	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/utils-go/crex"
)

const (

	// Hostname used for the single local compute unit in a blueprint plan.
	localComputeHost = "localhost"

	// Identifier for the default compute unit created in a blueprint plan.
	defaultComputeID = "default"
)

// Builds a Crucible blueprint resource from its configuration.
type Builder struct {
	src     hub.Source       // Registry access for resolving service references.
	env     string           // Environment selector for the blueprint.
	compute manifest.Compute // Compute instance assigned to the plan's default compute unit.
}

// Returns a new Builder.
//
// source provides registry access for resolving service references. env
// selects the blueprint environment to resolve. compute is the default compute
// instance assigned to the plan; when its type is empty it defaults to the
// local machine provider on the loopback host.
func NewBuilder(src hub.Source, env string, compute manifest.Compute) *Builder {
	if compute.Type == "" {
		compute = defaultCompute()
	}
	return &Builder{src: src, env: env, compute: compute}
}

// Builds the blueprint described by cfg.
//
// Writes plan.yaml to the output directory. The plan includes references to
// all services and their resolved manifests, but does not include the service
// or runtime configs themselves. Those are expected to be pulled separately by
// the deployer when needed.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Blueprint, output string) error {
	p, err := resolvePlan(ctx, cfg, b.env, b.src, b.compute)
	if err != nil {
		return err
	}
	if err := manifest.WritePlanAt(p, output); err != nil {
		return crex.SystemError("cannot write deployment plan", "failed to write the deployment plan to the build directory").
			Recoveryf("Make sure you have write access to %s, then try again.", output).
			Cause(err).
			Err()
	}
	return nil
}

// Produces the deployment plan for a blueprint without writing it to disk.
//
// The selected environment is located by envID, then every service is resolved
// against src: its manifest is pulled, its environment variables are validated,
// and a deployment entry is emitted. Resolved services are bin-packed onto
// compute units, and cloud network entries are derived from per-container
// network specs. computeHost is set as the Host of the single local compute
// unit. Returns the assembled plan, or an error from environment lookup or any
// service resolution.
func resolvePlan(ctx context.Context, cfg *manifest.Blueprint, envID string, src hub.Source, compute manifest.Compute) (*manifest.Plan, error) {
	env, err := findEnvironment(cfg, envID)
	if err != nil {
		return nil, err
	}

	p := newPlan(cfg, compute)
	results, err := resolveServices(ctx, cfg.Services, env, src)
	if err != nil {
		return nil, err
	}
	for _, r := range results {
		p.Services[r.serviceID] = r.target
		p.Containers[r.serviceID] = r.container
	}

	assignments := binPack(results, p.Infrastructure.Computes)
	addDeployments(p, env.ID, results, assignments)

	p.Environments[env.ID] = *env
	deriveNetworks(p)
	return p, nil
}

// Produces a new deployment plan with the default compute entry.
func newPlan(cfg *manifest.Blueprint, compute manifest.Compute) *manifest.Plan {
	return &manifest.Plan{
		Version: manifest.PlanVersion,
		Infrastructure: manifest.Infrastructure{
			Computes: map[string]manifest.Compute{defaultComputeID: compute},
		},
		Services:     make(map[string]string),
		Containers:   make(map[string]manifest.Container),
		Environments: make(map[string]manifest.Environment),
		Gateway:      cfg.Gateway,
	}
}

// Produces the default compute configuration for a local blueprint plan.
func defaultCompute() manifest.Compute {
	return manifest.Compute{
		Type: manifest.ComputeTypeLocal,
		Config: &manifest.ComputeLocal{
			Host: localComputeHost,
		},
	}
}

// Resolves all services in a blueprint against the selected environment.
func resolveServices(ctx context.Context, services []manifest.Ref, env *manifest.Environment, src hub.Source) ([]serviceResult, error) {
	results := make([]serviceResult, 0, len(services))
	for _, service := range services {
		r, err := planService(ctx, service, env, src)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// Appends deployment entries derived from the assigned compute units.
func addDeployments(p *manifest.Plan, envID string, results []serviceResult, assignments map[string]string) {
	for _, r := range results {
		computeID := assignments[r.serviceID]
		p.Deployments = append(p.Deployments, manifest.Deployment{
			Service:     r.serviceID,
			Container:   r.serviceID,
			Environment: envID,
			Compute:     computeID,
			Network:     computeID,
		})
	}
}

// Resolved planning data for a single service.
type serviceResult struct {
	serviceID string             // manifest service identifier
	target    string             // resolved reference string for the service image
	container manifest.Container // resolved container config
}

// Resolves a single service reference into a serviceResult.
//
// Pulls the service manifest and validates the environment variables.
func planService(ctx context.Context, service manifest.Ref, env *manifest.Environment, src hub.Source) (serviceResult, error) {
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

	if serviceCfg.OutputStage() == nil {
		return serviceResult{}, crex.Newf(ErrBuildPlan, "service %s has no output stage", service.ID)
	}

	return serviceResult{
		serviceID: service.ID,
		target:    ref.String(),
		container: manifest.Container{},
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
	return nil, crex.Newf(ErrBuildPlan, "environment %q not found", envID)
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
			return crex.Newf(ErrBuildPlan, "missing required variable %q", p.Name)
		}
	}
	return nil
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

// Pulls a service resource and extracts its manifest config.
func resolveService(ctx context.Context, id string, ref *reference.Reference, src hub.Source) (*manifest.Service, error) {
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
