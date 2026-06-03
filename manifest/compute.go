package manifest

import (
	"github.com/cruciblehq/crux/codec"
	"github.com/cruciblehq/crux/crex"
)

// Represents one compute unit in the infrastructure plan.
//
// A compute unit is an allocation of capacity for hosting services and can be
// backed by a variety of providers (AWS EC2, local machine, etc). The provider
// type determines the configuration fields required to provision and connect
// to the compute unit. The executor provisions the compute units and schedules
// services onto them based on the deployment specifications in the plan.
type Compute struct {

	// Provider type.
	//
	// Determines the type of [Compute.Config], which holds the configuration
	// fields specific to that provider. Supported providers are determined by
	// the compute package. Currently supported providers are "aws" for AWS EC2
	// instances and "local".
	Type string `codec:"type"`

	// Provider-specific configuration.
	//
	// The concrete type depends on [Compute.Type]: [*ComputeAWS] for "aws",
	// [*ComputeLocal] for "local".
	Config any `codec:"-"`

	// Compiled VM-level security model derived from service grants.
	//
	// Populated by the blueprint builder after bin-packing. Nil when the
	// compute unit has not yet had a policy derived for it. The provider
	// applies this policy at provisioning time.
	Policy *ComputeSecurityModel `codec:"policy,omitempty"`
}

// Validates the compute entry.
//
// The provider type is not validated here since other packages determined the
// supported types. The provider-specific configuration (Compute.Config) must
// implement the Validatable interface and is validated by calling its Validate
// method. The configuration must also be present (non-nil), since the provider
// type implies required configuration fields. If a policy is present it is
// also validated.
func (c *Compute) Validate() error {
	if c.Config == nil {
		return ErrMissingComputeProvider
	}
	v, ok := c.Config.(codec.Validatable)
	if !ok {
		return crex.Wrapf(ErrInvalidComputeType, "config type %T is not validatable", c.Config)
	}
	if err := v.Validate(); err != nil {
		return err
	}
	if c.Policy != nil {
		if err := c.Policy.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Decodes a raw parsed map into the compute entry.
//
// Reads the provider type first, then decodes the remaining fields into the
// concrete configuration type for that provider. The optional "policy" key is
// decoded into a [ComputeSecurityModel] if present.
func (c *Compute) Decode(raw any) error {
	src, ok := raw.(map[string]any)
	if !ok {
		return crex.Wrapf(ErrDecodeFailed, "compute: unexpected type %T", raw)
	}
	if err := codec.Field(src, c, "Type"); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}
	var target any
	switch c.Type {
	case "aws":
		target = &ComputeAWS{}
	case "local":
		target = &ComputeLocal{}
	default:
		return crex.Wrapf(ErrInvalidComputeType, "unknown provider %q", c.Type)
	}
	if err := codec.Decode(src, target); err != nil {
		return crex.Wrap(ErrDecodeFailed, err)
	}
	c.Config = target
	if rawPolicy, ok := src["policy"]; ok && rawPolicy != nil {
		policyMap, ok := rawPolicy.(map[string]any)
		if !ok {
			return crex.Wrap(ErrDecodeFailed, ErrInvalidComputeSecurityModel)
		}
		p := &ComputeSecurityModel{}
		if err := codec.Decode(policyMap, p); err != nil {
			return crex.Wrap(ErrDecodeFailed, err)
		}
		c.Policy = p
	}
	return nil
}

// Encodes the compute entry to a serializable map.
//
// Merges the provider configuration fields with the provider type key into a
// single flat map. The policy, if present, is encoded as a nested "policy" key.
func (c *Compute) Encode() (any, error) {
	var cfg map[string]any
	if c.Config == nil {
		cfg = map[string]any{"type": c.Type}
	} else {
		var err error
		cfg, err = encodeToMap(c.Config)
		if err != nil {
			return nil, err
		}
		cfg["type"] = c.Type
	}
	if c.Policy != nil {
		raw, err := encodeToMap(c.Policy)
		if err != nil {
			return nil, err
		}
		cfg["policy"] = raw
	}
	return cfg, nil
}
