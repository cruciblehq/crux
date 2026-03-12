package resource

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/cache"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
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
		return Source{}, errors.New("default registry is required")
	}
	if namespace == "" {
		return Source{}, errors.New("default namespace is required")
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

// Pulls a resource from the registry using the local cache.
//
// The reference is parsed and the resource is fetched (or retrieved from cache).
// Returns both the pull result and the fully resolved reference.
func (s Source) Pull(ctx context.Context, resourceType manifest.ResourceType, ref string) (*PullResult, *reference.Reference, error) {
	resolved, err := s.Parse(resourceType, ref)
	if err != nil {
		return nil, nil, err
	}

	result, err := pull(ctx, resolved)
	if err != nil {
		return nil, nil, err
	}

	return result, resolved, nil
}

// Resolves a resource reference string to a local file path.
//
// The reference is parsed, defaults from this Source are applied for any
// missing registry or namespace, and then the resource is pulled from the
// registry (with caching) and extracted. The returned path points to the
// image file inside the extracted archive. If the resource is already cached
// and extracted, no download occurs.
func (s Source) Resolve(ctx context.Context, resourceType manifest.ResourceType, ref string) (string, *reference.Reference, error) {
	result, resolved, err := s.Pull(ctx, resourceType, ref)
	if err != nil {
		return "", nil, err
	}

	localCache, err := cache.Open()
	if err != nil {
		return "", nil, crex.Wrap(ErrSourceResolve, err)
	}
	defer localCache.Close()

	extractDir, err := localCache.Extract(result.Namespace, result.Resource, result.Version)
	if err != nil {
		return "", nil, crex.Wrap(ErrSourceResolve, err)
	}

	path := filepath.Join(extractDir, manifest.ImageFile)
	return path, resolved, nil
}
