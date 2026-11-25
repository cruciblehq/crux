package blueprint

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/crex"
)

const (
	localComputeHost = "localhost" // Hostname of the single local machine in a blueprint plan.
	defaultComputeID = "default"   // Identifier of the default compute entry created in a blueprint plan.
)

// Builds a Crucible blueprint resource from its configuration.
type Builder struct {
	src     registry.Source  // Registry access for resolving service references.
	dir     string           // Base directory for relative service paths.
	env     string           // Environment selector for the blueprint.
	compute manifest.Compute // Compute entry used as the plan's default.
}

// Returns a new Builder.
//
// src provides registry access for resolving service references. dir is the
// directory that relative service paths are resolved against. env selects the
// blueprint environment to resolve. compute is the plan's default compute entry.
// When its type is empty, it defaults to the local provider on localhost.
func NewBuilder(src registry.Source, dir, env string, compute manifest.Compute) *Builder {
	if compute.Type == "" {
		compute = defaultCompute()
	}
	return &Builder{src: src, dir: dir, env: env, compute: compute}
}

// Builds the blueprint described by cfg and writes its plan to w.
//
// The plan references every service but does not include the service
// manifests or their runtime configs. The deployer pulls those separately
// when needed.
func (b *Builder) Build(ctx context.Context, cfg *manifest.Blueprint, w io.Writer) error {
	p, err := resolvePlan(ctx, cfg, b.env, b.src, b.dir, b.compute)
	if err != nil {
		return err
	}
	return manifest.EncodePlan(w, p)
}

// Produces the deployment plan for a blueprint without writing it to disk.
//
// The selected environment is located by envID, then every service is resolved.
// Its manifest is pulled from src, or read from its path resolved against dir.
// Its environment variables are validated, and a deployment entry is emitted.
// compute becomes the plan's default entry, and every service is assigned to a
// compute entry ([binPack]). Network entries of the infrastructure are derived
// from the network specs of the containers. Returns the assembled plan, or an
// error from environment lookup or any service resolution.
func resolvePlan(ctx context.Context, cfg *manifest.Blueprint, envID string, src registry.Source, dir string, compute manifest.Compute) (*manifest.Plan, error) {
	env, err := findEnvironment(cfg, envID)
	if err != nil {
		return nil, err
	}

	p := newPlan(cfg, compute)
	results, err := resolveServices(ctx, cfg.Services, env, src, dir)
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
func resolveServices(ctx context.Context, services []manifest.Ref, env *manifest.Environment, src registry.Source, dir string) ([]serviceResult, error) {
	results := make([]serviceResult, 0, len(services))
	for _, service := range services {
		r, err := planService(ctx, service, env, src, dir)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// Appends deployment entries derived from the assigned compute entries.
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
	serviceID string             // Service identifier in the blueprint.
	target    string             // Resolved reference of the service.
	container manifest.Container // Resolved container config.
}

// Resolves a single service reference into a serviceResult.
//
// Loads the service manifest, validates the environment variables, and checks
// that the service has an output stage.
func planService(ctx context.Context, service manifest.Ref, env *manifest.Environment, src registry.Source, dir string) (serviceResult, error) {
	target, serviceCfg, err := resolveService(ctx, service, src, dir)
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
		target:    target,
		container: manifest.Container{},
	}, nil
}

// Locates the selected environment in the blueprint.
//
// Returns an error if no environment with that ID is declared. When the
// blueprint declares no environments, returns an empty environment with that
// ID.
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

// Assigns services to compute entries.
//
// All services go to one compute entry, picked in map order, which is
// arbitrary when computes holds more than one. This is a placeholder. A real
// packer can replace it without changing callers.
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

// Loads a service's manifest config and returns it with the service's plan
// target.
//
// A service named by path is read from its directory, resolved against dir
// when relative. Its target is that absolute directory. A service named by
// reference is pulled from src. Its target is the parsed reference.
func resolveService(ctx context.Context, service manifest.Ref, src registry.Source, dir string) (string, *manifest.Service, error) {
	if service.Path != "" {
		path := service.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}
		cfg, err := readService(service.ID, path)
		if err != nil {
			return "", nil, err
		}
		return path, cfg, nil
	}

	ref, err := src.Parse(string(manifest.TypeService), service.Ref)
	if err != nil {
		return "", nil, errService(service.ID, err)
	}
	result, err := src.Pull(ctx, ref)
	if err != nil {
		return "", nil, errService(service.ID, err)
	}
	cfg, err := readService(service.ID, result.Extracted)
	if err != nil {
		return "", nil, err
	}
	return ref.String(), cfg, nil
}

// Decodes the service manifest held in dir for the service with the given id.
func readService(id, dir string) (*manifest.Service, error) {
	f, err := os.Open(filepath.Join(dir, manifest.ManifestFile))
	if err != nil {
		return nil, errServiceAt(id, dir, err)
	}
	defer f.Close()
	m, err := manifest.Decode(f)
	if err != nil {
		return nil, errServiceAt(id, dir, err)
	}
	cfg, err := manifest.As[*manifest.Service](m)
	if err != nil {
		return nil, errServiceAt(id, dir, err)
	}
	return cfg, nil
}
