package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
	"github.com/cruciblehq/crux/internal/reference"
)

// Configures the default registry and namespace used when resolving resource
// references that do not specify their own.
type Source struct {
	Registry  string // Registry URL when not specified in the reference.
	Namespace string // Namespace when not specified in the reference.
}

// Creates a new [Source] with the given registry and namespace.
//
// Both parameters are required. Returns an error if either is empty.
func NewSource(registry, namespace string) (Source, error) {
	if registry == "" {
		return Source{}, crex.Wrap(ErrMissingOption, ErrMissingRegistry)
	}
	if namespace == "" {
		return Source{}, crex.Wrap(ErrMissingOption, ErrMissingNamespace)
	}
	return Source{Registry: registry, Namespace: namespace}, nil
}

// Parses a resource reference string and applies this Source's defaults for
// any missing registry or namespace.
func (s Source) Parse(resourceType manifest.ResourceType, ref string) (*reference.Reference, error) {
	parsed, err := reference.Parse(ref, string(resourceType))
	if err != nil {
		return nil, err
	}
	return parsed.WithDefaults(s.Registry, s.Namespace), nil
}

// Pulls a resource from the registry and extracts it locally.
func (s Source) Pull(ctx context.Context, ref *reference.Reference) (*PullResult, error) {
	return pull(ctx, ref)
}

// Pushes a resource package to the Hub registry.
func (s Source) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, s.Registry, m, packagePath)
}
