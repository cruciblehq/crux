package resource

import (
	"errors"

	"github.com/cruciblehq/spec/reference"
)

// Default registry and namespace applied when references or identifiers
// do not specify their own.
//
// Builders and push operations store Defaults rather than [reference.Options]
// so the package does not misuse a parsing-time type as ambient configuration.
// Convert to the appropriate reference type using [Defaults.IdentifierOptions]
// or [Defaults.ReferenceOptions].
type Defaults struct {
	Registry  string // Registry authority when not specified.
	Namespace string // Namespace when not specified.
}

// Creates a new [Defaults] with the given registry and namespace.
//
// Both parameters are required. Returns an error if either is empty.
func NewDefaults(registry, namespace string) (Defaults, error) {
	if registry == "" {
		return Defaults{}, errors.New("default registry is required")
	}
	if namespace == "" {
		return Defaults{}, errors.New("default namespace is required")
	}
	return Defaults{Registry: registry, Namespace: namespace}, nil
}

// Returns [reference.IdentifierOptions] populated from these defaults.
func (d Defaults) IdentifierOptions() reference.IdentifierOptions {
	return reference.IdentifierOptions{
		DefaultRegistry:  d.Registry,
		DefaultNamespace: d.Namespace,
	}
}

// Returns [reference.Options] populated from these defaults.
func (d Defaults) ReferenceOptions() reference.Options {
	return reference.Options{
		IdentifierOptions: d.IdentifierOptions(),
	}
}
