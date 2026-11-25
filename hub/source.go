package hub

import (
	"context"
	"io"

	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/utils-go/crex"
)

// Configures the default registry and namespace used when resolving references.
//
// When a reference string omits the registry host or the namespace component,
// the corresponding default is applied by [Source.Parse] before the reference
// is passed to any pull or push operation.
type Source struct {
	Registry   string // Default registry URL.
	Namespace  string // Default namespace.
	Credential string // Credential for Registry; empty for none.
}

// Creates a new [Source] with the given registry and namespace.
//
// Both parameters are required. Returns an error wrapping [ErrMissingOption]
// if either is empty.
func NewSource(registry, namespace string) (Source, error) {
	if registry == "" {
		return Source{}, crex.ProgrammingError("missing registry", "a registry URL is required").
			Cause(crex.Wrap(ErrMissingOption, ErrMissingRegistry)).
			Err()
	}
	if namespace == "" {
		return Source{}, crex.ProgrammingError("missing namespace", "a namespace is required").
			Cause(crex.Wrap(ErrMissingOption, ErrMissingNamespace)).
			Err()
	}
	return Source{Registry: registry, Namespace: namespace}, nil
}

// Parses a resource reference string for the given type.
//
// Applies the default registry and namespace from [Source] when those
// components are absent from the reference string.
func (s Source) Parse(resourceType, ref string) (*reference.Reference, error) {
	parsed, err := reference.Parse(ref, resourceType)
	if err != nil {
		return nil, err
	}
	return parsed.WithDefaults(s.Registry, s.Namespace), nil
}

// Pulls a resource from the registry and extracts it locally.
//
// Checks the local cache first. Only downloads on cache miss or if the cached
// version is invalid or its digest does not match. Returns a [PullResult]
// containing the local extraction directory, digest, and version metadata.
// The credential is sent only when ref points at the source's registry.
func (s Source) Pull(ctx context.Context, ref *reference.Reference) (*PullResult, error) {
	credential := ""
	if ref.Registry() == s.Registry {
		credential = s.Credential
	}
	return pull(ctx, ref, credential)
}

// Pushes a resource package to the registry.
//
// Creates the resource and the version if they do not exist yet, then uploads
// archive. Pushing a private version again replaces its archive. Returns
// [ErrVersionPublic] if the version is public, or [ErrVersionSuperseded] if a
// public version with the same major and minor is higher.
func (s Source) Push(ctx context.Context, name, resourceType, version string, archive io.ReadSeeker) error {
	return push(ctx, s.Registry, s.Credential, name, resourceType, version, archive)
}
