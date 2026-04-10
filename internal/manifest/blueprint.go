package manifest

import "github.com/cruciblehq/crex"

// Holds configuration specific to blueprint resources.
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
