package manifest

import (
	"github.com/cruciblehq/crux/crex"
)

// A blueprint model.
//
// A blueprint declares which services should be deployed together and how
// they are composed and exposed. Building a blueprint resolves service
// references and affordances, producing a deployment plan as output.
type Blueprint struct {

	// Services to deploy.
	//
	// Each entry names a service from the registry. Services carry only
	// an ID and a reference; affordances and configuration come from the
	// service manifest fetched during the build.
	Services []Ref `codec:"services"`

	// Gateway routing configuration.
	//
	// Maps URL path patterns to service IDs. Services without a route
	// still run but do not receive external traffic.
	Gateway Gateway `codec:"gateway"`

	// Named environment variable sets.
	//
	// Each environment provides concrete values for the config/env and
	// config/secret affordances declared by services. Which environment
	// to use is selected at build time.
	Environments []Environment `codec:"environments,omitempty"`
}

// Returns the service ref with the given ID.
//
// The service must be declared in the blueprint, otherwise nil is returned.
// The returned ref is a pointer to the entry in the blueprint's service list,
// so changes to it will be reflected in the blueprint.
func (b *Blueprint) Service(id string) *Ref {
	for i := range b.Services {
		if b.Services[i].ID == id {
			return &b.Services[i]
		}
	}
	return nil
}

// Appends a service reference to the blueprint's service list.
//
// Adding a service does not start new instances, it only affects the blueprint
// declaration. After a successful addition, the deployer is expected to build
// a new plan and deploy it. The new plan will include the new service and the
// deployer can choose how to reconcile the diff between the previous and new
// plan. The ID must be unique within the blueprint. Returns [ErrServiceExists]
// if a service with the same ID is already registered.
func (b *Blueprint) AddService(ref Ref) error {
	if b.Service(ref.ID) != nil {
		return crex.Wrapf(ErrServiceExists, "%s", ref.ID)
	}
	b.Services = append(b.Services, ref)
	return nil
}

// Removes the service with the given ID from the blueprint.
//
// Removing a service does not destroy a running instance, it only affects the
// blueprint declaration. After a successful removal, the deployer is expected
// to build a new plan and deploy it. The new plan will no longer include the
// removed service and the deployer can choose how to reconcile the diff between
// the previous and new plan. The ID must exist in the blueprint, otherwise
// [ErrServiceNotFound] is returned.
func (b *Blueprint) RemoveService(id string) error {
	for i, s := range b.Services {
		if s.ID == id {
			b.Services = append(b.Services[:i], b.Services[i+1:]...)
			return nil
		}
	}
	return crex.Wrapf(ErrServiceNotFound, "%s", id)
}

// Returns the environment with the given ID.
//
// The environment must be declared in the blueprint, otherwise nil is returned.
// The returned environment is a pointer to the entry in the blueprint's
// environment list, so changes to it will be reflected in the blueprint.
func (b *Blueprint) Environment(id string) *Environment {
	for i := range b.Environments {
		if b.Environments[i].ID == id {
			return &b.Environments[i]
		}
	}
	return nil
}

// Appends an environment to the blueprint.
//
// Environment IDs are selected at build time to provide concrete values for
// config/env and config/secret affordances. The ID must be unique within
// the blueprint. Returns [ErrEnvironmentExists] if an environment with the
// same ID is already declared.
func (b *Blueprint) AddEnvironment(env Environment) error {
	if b.Environment(env.ID) != nil {
		return crex.Wrapf(ErrEnvironmentExists, "%s", env.ID)
	}
	b.Environments = append(b.Environments, env)
	return nil
}

// Removes the environment with the given ID from the blueprint.
//
// Removal only affects the blueprint declaration; any plan built against the
// removed environment remains valid until rebuilt. Returns [ErrEnvironmentNotFound]
// if no environment with that ID is declared.
func (b *Blueprint) RemoveEnvironment(id string) error {
	for i, e := range b.Environments {
		if e.ID == id {
			b.Environments = append(b.Environments[:i], b.Environments[i+1:]...)
			return nil
		}
	}
	return crex.Wrapf(ErrEnvironmentNotFound, "%s", id)
}

// Validates the blueprint configuration.
//
// Service IDs must be unique. Every route must reference an existing service.
// Route patterns must be unique. Environment IDs must be unique.
func (b *Blueprint) Validate() error {
	ids, err := b.validateServices()
	if err != nil {
		return err
	}

	if err := b.validateRoutes(ids); err != nil {
		return err
	}

	return b.validateEnvironments()
}

// Validates all service entries.
//
// Each service must have a valid ref and a non-empty, unique ID.
func (b *Blueprint) validateServices() (map[string]struct{}, error) {
	ids := make(map[string]struct{}, len(b.Services))
	for i := range b.Services {
		if err := b.Services[i].Validate(); err != nil {
			return nil, crex.Wrap(ErrInvalidBlueprint, err)
		}
		if b.Services[i].ID == "" {
			return nil, crex.Wrap(ErrInvalidBlueprint, ErrMissingServiceID)
		}
		if _, exists := ids[b.Services[i].ID]; exists {
			return nil, crex.Wrap(ErrInvalidBlueprint, ErrDuplicateServiceID)
		}
		ids[b.Services[i].ID] = struct{}{}
	}
	return ids, nil
}

// Validates the gateway and its routes.
//
// The gateway must be individually valid and every route must reference
// a service declared in the blueprint.
func (b *Blueprint) validateRoutes(ids map[string]struct{}) error {
	if err := b.Gateway.Validate(); err != nil {
		return crex.Wrap(ErrInvalidBlueprint, err)
	}
	for _, route := range b.Gateway.Routes {
		if _, exists := ids[route.Service]; !exists {
			return crex.Wrap(ErrInvalidBlueprint, ErrRouteServiceNotFound)
		}
	}
	return nil
}

// Validates all environment entries.
//
// Each environment must be individually valid and environment IDs must be unique.
func (b *Blueprint) validateEnvironments() error {
	envIDs := make(map[string]struct{}, len(b.Environments))
	for i := range b.Environments {
		if err := b.Environments[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidBlueprint, err)
		}
		if _, exists := envIDs[b.Environments[i].ID]; exists {
			return crex.Wrap(ErrInvalidBlueprint, ErrDuplicateEnvironmentID)
		}
		envIDs[b.Environments[i].ID] = struct{}{}
	}
	return nil
}
