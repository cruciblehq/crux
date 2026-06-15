package manifest

import "github.com/cruciblehq/crux/crex"

// Declares the parameters a resource accepts.
//
// Params lists the named parameters. Default optionally names one of them
// as the recipient of scalar values. When a caller passes a plain value
// instead of a named argument map, it is assigned to the default parameter.
// If Default is set, it must reference an existing param. Zero value means
// no parameters.
type Schema struct {

	// Name of the parameter that receives scalar values.
	//
	// When a caller passes a plain value instead of a named argument map,
	// the value is assigned to this parameter. A parameter with the same
	// name must exist in Params. Zero value means no default.
	Default string `codec:"default,omitempty"`

	// Named parameters accepted by the resource.
	//
	// Each param must have a unique name. The Default field, when set, must
	// reference one of these names. Zero value means no parameters.
	Params []Param `codec:"params,omitempty"`
}

// Validates the schema.
//
// All parameters must be valid, names must be unique, and if Default is set
// it must reference an existing param.
func (s *Schema) Validate() error {
	seen := make(map[string]bool, len(s.Params))

	for i := range s.Params {
		p := &s.Params[i]
		if err := p.Validate(); err != nil {
			return err
		}

		if seen[p.Name] {
			return crex.Tag(crex.Newf(ErrInvalidParam, "param %q is a duplicate", p.Name), ErrDuplicateParamName)
		}
		seen[p.Name] = true
	}

	if s.Default != "" && !isValidName(s.Default) {
		return crex.Tag(crex.Newf(ErrInvalidParam, "default %q has an invalid name", s.Default), ErrInvalidParamName)
	}

	if s.Default != "" && !seen[s.Default] {
		return crex.Tag(crex.Newf(ErrInvalidParam, "default %q is not in the schema", s.Default), ErrDefaultNotInSchema)
	}

	return nil
}
