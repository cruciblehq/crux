package registry

import (
	"context"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/reference"
)

// Configures the default registry and namespace used when resolving references.
//
// When a reference string omits the registry host or the namespace component,
// the corresponding default is applied by [Source.Parse] before the reference
// is passed to any pull or push operation.
type Source struct {
	Registry  string // Default registry URL.
	Namespace string // Default namespace.
}

// Creates a new [Source] with the given registry and namespace.
//
// Both parameters are required. Returns an error wrapping [ErrMissingOption]
// if either is empty.
func NewSource(registry, namespace string) (Source, error) {
	if registry == "" {
		return Source{}, crex.Wrap(ErrMissingOption, ErrMissingRegistry)
	}
	if namespace == "" {
		return Source{}, crex.Wrap(ErrMissingOption, ErrMissingNamespace)
	}
	return Source{Registry: registry, Namespace: namespace}, nil
}

// Parses a resource reference string for the given type.
//
// Applies the default registry and namespace from [Source] when those
// components are absent from the reference string.
func (s Source) Parse(resourceType string, ref string) (*reference.Reference, error) {
	parsed, err := reference.Parse(ref, resourceType)
	if err != nil {
		return nil, err
	}
	return parsed.WithDefaults(s.Registry, s.Namespace), nil
}

// Pulls a resource from the registry and extracts it locally.
//
// Checks the local cache first; only downloads on cache miss or if the cached
// version is invalid or its digest does not match. Returns a [PullResult]
// containing the local extraction directory, digest, and version metadata.
func (s Source) Pull(ctx context.Context, ref *reference.Reference) (*PullResult, error) {
	return pull(ctx, ref)
}

// Pushes a resource package to the registry.
//
// Creates the resource and version in the registry if they do not already
// exist, then uploads the archive at packagePath.
func (s Source) Push(ctx context.Context, name, resourceType, version, packagePath string) error {
	return push(ctx, s.Registry, name, resourceType, version, packagePath)
}
