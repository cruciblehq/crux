package store

import (
	"context"

	"github.com/cruciblehq/crux/pkg/reference"
)

// Defines methods for interacting with a remote registry.
//
// Implementations of this interface are responsible for communicating with
// remote Crucible registries to query available resources and fetch archives.
// Each implementation may use different protocols or APIs depending on the
// registry it targets.
//
// The default implementation communicates with the official Crucible registry
// at https://registry.crucible.net. Use [NewRemote] to create an instance.
type Remote interface {

	// Returns namespace metadata.
	//
	// If the request includes an ETag from a previous response and the
	// namespace has not changed, Data will be nil and only ETag is set.
	Namespace(ctx context.Context, req *NamespaceRequest) (*NamespaceResponse, error)

	// Returns resource metadata.
	//
	// The versions list is ordered by semantic version in descending order
	// (newest first). If the request includes an ETag from a previous response
	// and the resource has not changed, Data will be nil and only ETag is set.
	Resource(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error)

	// Fetches a resource archive by version.
	//
	// Returns the raw archive data. If the request includes an ETag from a
	// previous response and the archive has not changed, Data will be nil and
	// only ETag is set.
	Fetch(ctx context.Context, req *FetchRequest) (*FetchResponse, error)

	// Fetches a resource archive by channel.
	//
	// Returns the raw archive data for the latest version on that channel. The
	// resolved version is included in the response. If the request includes
	// an ETag from a previous response and the archive has not changed, Data
	// will be nil and only ETag and Version are set.
	Consume(ctx context.Context, req *ConsumeRequest) (*ConsumeResponse, error)
}

// Request for namespace metadata.
//
// The Namespace field specifies the namespace identifier to query. The ETag
// field may contain a previously received ETag for cache validation. If not
// set, the full namespace data will be returned.
type NamespaceRequest struct {
	Namespace string // Namespace identifier.
	ETag      string // ETag for cache validation (may be empty).
}

// Response containing namespace metadata.
//
// The Data field contains the namespace information. If the namespace has not
// changed since the provided ETag, Data will be nil. The ETag field contains
// the current ETag for the namespace.
type NamespaceResponse struct {
	Info *NamespaceInfo // Namespace information (nil if not modified).
	ETag string         // ETag for cache validation.
}

// Request for resource metadata.
//
// The Identifier field specifies the resource identifier to query. The ETag
// field may contain a previously received ETag for cache validation. If not
// set, the full resource data will be returned.
type ResourceRequest struct {
	Identifier reference.Identifier // Resource identifier.
	ETag       string               // ETag for cache validation (may be empty).
}

// Response containing resource metadata.
//
// The Data field contains the resource information. If the resource has not
// changed since the provided ETag, Data will be nil. The ETag field contains
// the current ETag for the resource.
type ResourceResponse struct {
	Info *ResourceInfo // Resource information (nil if not modified).
	ETag string        // ETag for cache validation.
}

// Request for a resource archive by version.
//
// The Identifier field specifies the resource identifier to fetch and the
// Version field the specific version to fetch. The ETag field may contain
// a previously received ETag for cache validation. If not set, the full
// archive data will be returned.
type FetchRequest struct {
	Identifier reference.Identifier // Resource identifier.
	Version    reference.Version    // Specific version to fetch.
	ETag       string               // ETag for cache validation (may be empty).
}

// Response containing a resource archive.
//
// The Data field contains the raw archive data. If the archive has not changed
// since the provided ETag, Data will be nil. The ETag field contains the
// current ETag for the archive.
type FetchResponse struct {
	Data []byte // Raw archive data (nil if not modified).
	ETag string // ETag for cache validation.
}

// Request for a resource archive by channel.
//
// The Identifier field specifies the resource identifier to fetch and the
// Channel field the channel to fetch (e.g., "stable", "latest"). The ETag
// field may contain a previously received ETag for cache validation. If not
// set, the full archive data will be returned.
type ConsumeRequest struct {
	Identifier reference.Identifier // Resource identifier.
	Channel    string               // Channel to fetch.
	ETag       string               // ETag for cache validation (may be empty).
}

// Response containing a resource archive fetched by channel.
//
// The Version field contains the resolved version for the specified channel.
// The Data field contains the raw archive data. If the archive has not changed
// since the provided ETag, Data will be nil. The ETag field contains the
// current ETag for the archive.
type ConsumeResponse struct {
	Version reference.Version // Resolved version for this channel.
	Data    []byte            // Raw archive data (nil if not modified).
	ETag    string            // ETag for cache validation.
}
