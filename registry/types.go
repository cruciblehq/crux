package registry

import "fmt"

// String identifier for HTTP Content-Type and Accept headers.
//
// Defines vendor-specific media types for the Crucible registry API following
// the pattern application/vnd.crucible.{name}.v0. Used in Content-Type headers
// for request bodies and Accept headers for response format negotiation.
type MediaType string

const (
	MediaTypeError         MediaType = "application/vnd.crucible.error.v0"          // Error responses with codes and messages.
	MediaTypeNamespaceInfo MediaType = "application/vnd.crucible.namespace-info.v0" // Namespace create/update requests.
	MediaTypeNamespace     MediaType = "application/vnd.crucible.namespace.v0"      // Complete namespace with resource summaries.
	MediaTypeNamespaceList MediaType = "application/vnd.crucible.namespace-list.v0" // Collection of namespace summaries.
	MediaTypeResourceInfo  MediaType = "application/vnd.crucible.resource-info.v0"  // Resource create/update requests.
	MediaTypeResource      MediaType = "application/vnd.crucible.resource.v0"       // Complete resource with version/channel summaries.
	MediaTypeResourceList  MediaType = "application/vnd.crucible.resource-list.v0"  // Collection of resource summaries.
	MediaTypeVersionInfo   MediaType = "application/vnd.crucible.version-info.v0"   // Version create/update requests.
	MediaTypeVersion       MediaType = "application/vnd.crucible.version.v0"        // Complete version with archive details.
	MediaTypeVersionList   MediaType = "application/vnd.crucible.version-list.v0"   // Collection of version summaries.
	MediaTypeChannelInfo   MediaType = "application/vnd.crucible.channel-info.v0"   // Channel create/update requests.
	MediaTypeChannel       MediaType = "application/vnd.crucible.channel.v0"        // Complete channel with full version object.
	MediaTypeChannelList   MediaType = "application/vnd.crucible.channel-list.v0"   // Collection of channel summaries.
	MediaTypeArchive       MediaType = "application/vnd.crucible.archive.v0"        // Binary archive data (tar.zst format).
)

// Platform-specific error code for machine-readable error classification.
//
// Provides granular error information beyond HTTP status codes, enabling
// clients to implement specific error handling logic. Used in the Code field
// of Error responses.
type ErrorCode string

const (
	ErrorCodeBadRequest           ErrorCode = "bad_request"                     // Request validation failed (malformed body, invalid fields).
	ErrorCodeNotFound             ErrorCode = "not_found"                       // Requested namespace, resource, version, or channel does not exist.
	ErrorCodeNamespaceExists      ErrorCode = "namespace_exists"                // Cannot create namespace - name already in use.
	ErrorCodeNamespaceNotEmpty    ErrorCode = "namespace_not_empty"             // Cannot delete namespace - contains resources.
	ErrorCodeResourceExists       ErrorCode = "resource_exists"                 // Cannot create resource - name already in use within namespace.
	ErrorCodeResourceHasPublished ErrorCode = "resource_has_published_versions" // Cannot delete resource - contains published versions.
	ErrorCodeVersionExists        ErrorCode = "version_exists"                  // Cannot create version - version string already in use.
	ErrorCodeVersionPublished     ErrorCode = "version_published"               // Cannot modify or delete version - already published and immutable.
	ErrorCodeChannelExists        ErrorCode = "channel_exists"                  // Cannot create channel - name already in use.
	ErrorCodePreconditionFailed   ErrorCode = "precondition_failed"             // Request precondition not met (e.g., If-Match header mismatch).
	ErrorCodeInternalError        ErrorCode = "internal_error"                  // Unexpected server error occurred.
)

// Error response from the registry API.
//
// Provides both machine-readable error classification through the Code field
// and human-readable context through the Message field. The media type is
// [MediaTypeError].
type Error struct {
	Code    ErrorCode `json:"code"`    // Error code (see [ErrorCode]).
	Message string    `json:"message"` // Human-readable error description.
}

// Implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Mutable properties of a namespace for creation or update.
//
// Used as the request body for namespace creation and update operations. The
// name is the unique identifier for the namespace and cannot be changed once
// created. For update requests, Name must match the URL path parameter or
// update context. Contains only user-modifiable fields; system-managed fields
// are set by the server and appear only in response types. The media type is
// [MediaTypeNamespaceInfo].
type NamespaceInfo struct {
	Name        string `json:"name"`        // Namespace name.
	Description string `json:"description"` // Human-readable description.
}

// Lightweight namespace representation for listings.
//
// Provides namespace metadata without nested resource details. Used in list
// responses to keep payloads compact. Includes read-only fields like creation
// timestamps and statistics that are not present in [NamespaceInfo].
type NamespaceSummary struct {
	Name          string `json:"name"`          // Namespace name.
	Description   string `json:"description"`   // Human-readable description.
	ResourceCount int    `json:"resourceCount"` // Number of resources in this namespace.
	CreatedAt     int64  `json:"createdAt"`     // When the namespace was created.
	UpdatedAt     int64  `json:"updatedAt"`     // When the namespace was last updated.
}

// Complete namespace with its resource listings.
//
// Serves as the organizational unit for grouping resources. The resources list
// contains lightweight [ResourceSummary] entries without full version and channel
// details. For complete resource information, fetch individual resources. The
// media type is [MediaTypeNamespace].
type Namespace struct {
	Name        string            `json:"name"`        // Namespace name.
	Description string            `json:"description"` // Human-readable description.
	Resources   []ResourceSummary `json:"resources"`   // List of resources (summary form).
	CreatedAt   int64             `json:"createdAt"`   // When the namespace was created.
	UpdatedAt   int64             `json:"updatedAt"`   // When the namespace was last updated.
}

// Collection of namespaces.
//
// Namespaces may be empty if the registry contains no namespaces. The media
// type is [MediaTypeNamespaceList].
type NamespaceList struct {
	Namespaces []NamespaceSummary `json:"namespaces"` // List of namespaces.
}

// Mutable properties of a resource for creation or update.
//
// Used as the request body for resource creation and update operations. For
// update requests, the name field must match the URL path parameter or update
// context. Contains only user-modifiable fields. The media type is
// [MediaTypeResourceInfo].
type ResourceInfo struct {
	Name        string `json:"name"`        // Resource name.
	Type        string `json:"type"`        // Resource type (e.g., "widget", "service").
	Description string `json:"description"` // Human-readable description.
}

// Lightweight resource representation for listings.
//
// Provides resource metadata without nested version and channel details. Used
// in namespace listings and resource lists to keep payloads compact. Includes
// read-only fields like timestamps, statistics, and latest version information
// to help with navigation decisions.
type ResourceSummary struct {
	Name          string  `json:"name"`          // Resource name.
	Type          string  `json:"type"`          // Resource type (e.g., "widget", "service").
	Description   string  `json:"description"`   // Human-readable description.
	LatestVersion *string `json:"latestVersion"` // Most recent version string (null if no versions).
	VersionCount  int     `json:"versionCount"`  // Number of versions for this resource.
	ChannelCount  int     `json:"channelCount"`  // Number of channels for this resource.
	CreatedAt     int64   `json:"createdAt"`     // When the resource was created.
	UpdatedAt     int64   `json:"updatedAt"`     // When the resource was last updated.
}

// Complete resource with all its versions and channels.
//
// Provides comprehensive resource information including metadata, versions, and
// channels. The versions and channels lists contain lightweight summary entries
// without full archive details. For complete version information, fetch version
// details. Includes scoping information to identify the resource's location. The
// media type is [MediaTypeResource].
type Resource struct {
	Namespace   string           `json:"namespace"`   // Namespace this resource belongs to.
	Name        string           `json:"name"`        // Resource name.
	Type        string           `json:"type"`        // Resource type (e.g., "widget", "service").
	Description string           `json:"description"` // Human-readable description.
	Versions    []VersionSummary `json:"versions"`    // List of versions (summary form).
	Channels    []ChannelSummary `json:"channels"`    // List of channels (summary form).
	CreatedAt   int64            `json:"createdAt"`   // When the resource was created.
	UpdatedAt   int64            `json:"updatedAt"`   // When the resource was last updated.
}

// Collection of resources.
//
// The media type is [MediaTypeResourceList].
type ResourceList struct {
	Resources []ResourceSummary `json:"resources"` // List of resources.
}

// Mutable properties of a version for creation or update.
//
// Used as the request body for version creation and update operations. For
// update requests, the version field must match the URL path parameter or
// update context. Contains only user-modifiable fields. The media type is
// [MediaTypeVersionInfo].
type VersionInfo struct {
	String string `json:"string"` // Version string (e.g., "1.0.0").
}

// Lightweight version representation for listings.
//
// Provides version metadata without full archive details. Used in resource
// listings and version lists to keep payloads compact. Includes read-only
// fields like publication status and timestamps.
type VersionSummary struct {
	String    string `json:"string"`    // Version string (e.g., "1.0.0").
	CreatedAt int64  `json:"createdAt"` // When the version was created.
	UpdatedAt int64  `json:"updatedAt"` // When the version was last updated.
}

// Complete version with archive details and publication status.
//
// Tracks both metadata (always mutable) and archive state (immutable after
// publication). The archive, size, and digest fields are null before archive
// upload and populated afterward. The publishedAt field is null for unpublished
// versions and contains the publication timestamp when published. Unpublished
// versions support archive replacement for iterative development, while
// published versions ensure immutability for stable dependency resolution.
// Version metadata updates remain allowed even after publication. Includes
// scoping information to identify the version's location. The media type is
// [MediaTypeVersion].
type Version struct {
	Namespace string  `json:"namespace"` // Namespace this version belongs to.
	Resource  string  `json:"resource"`  // Resource this version belongs to.
	String    string  `json:"string"`    // Version string (e.g., "1.0.0").
	Archive   *string `json:"archive"`   // Download URL or null if not uploaded.
	Size      *int64  `json:"size"`      // Archive size in bytes (null if not uploaded).
	Digest    *string `json:"digest"`    // Archive digest (e.g., "sha256:abc...", null if not uploaded).
	CreatedAt int64   `json:"createdAt"` // When the version was created.
	UpdatedAt int64   `json:"updatedAt"` // When the version was last updated.
}

// Collection of versions for a resource.
//
// The media type is [MediaTypeVersionList].
type VersionList struct {
	Versions []VersionSummary `json:"versions"` // List of versions.
}

// Mutable properties of a channel for creation or update.
//
// Used as the request body for channel creation and update operations. For
// update requests, the name field must match the URL path parameter or update
// context. The version field is a simple string reference to an existing version;
// changing this pointer updates where the channel points. Contains only user-
// modifiable fields. The media type is [MediaTypeChannelInfo].
type ChannelInfo struct {
	Name        string `json:"name"`        // Channel name.
	Version     string `json:"version"`     // Version this channel points to.
	Description string `json:"description"` // Human-readable description.
}

// Lightweight channel representation for listings.
//
// Provides channel metadata with a version string reference. Used in resource
// listings and channel lists to keep payloads compact. Includes read-only
// fields like timestamps.
type ChannelSummary struct {
	Name        string `json:"name"`        // Channel name.
	Version     string `json:"version"`     // Version this channel points to.
	Description string `json:"description"` // Human-readable description.
	CreatedAt   int64  `json:"createdAt"`   // When the channel was created.
	UpdatedAt   int64  `json:"updatedAt"`   // When the channel was last updated.
}

// Mutable pointer to a version with complete version details.
//
// Provides a named reference that can be updated to point to different versions
// over time, primarily supporting QA/testing workflows. The embedded Version
// object provides full details about the currently targeted version, including
// archive availability and publication status. Channels enable dynamic version
// references during development but are discouraged for production use where
// explicit version references ensure reproducibility. Includes scoping
// information to identify the channel's location. The media type is
// [MediaTypeChannel].
type Channel struct {
	Namespace   string  `json:"namespace"`   // Namespace this channel belongs to.
	Resource    string  `json:"resource"`    // Resource this channel belongs to.
	Name        string  `json:"name"`        // Channel name.
	Version     Version `json:"version"`     // Full version object this channel points to.
	Description string  `json:"description"` // Human-readable description.
	CreatedAt   int64   `json:"createdAt"`   // When the channel was created.
	UpdatedAt   int64   `json:"updatedAt"`   // When the channel was last updated.
}

// Collection of channels with their current version targets.
//
// The media type is [MediaTypeChannelList].
type ChannelList struct {
	Channels []ChannelSummary `json:"channels"` // List of channels.
}
