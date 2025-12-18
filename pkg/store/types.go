package store

import (
	"fmt"
	"time"

	"github.com/cruciblehq/crux/pkg/reference"
)

const (
	MediaTypeNamespace = "application/vnd.crucible.namespace.v0+json" // Media type for [NamespaceInfo].
	MediaTypeResource  = "application/vnd.crucible.resource.v0+json"  // Media type for [ResourceInfo].
	MediaTypeArchive   = "application/vnd.crucible.archive.v0+zstd"   // Media type for resource package archives.
	MediaTypeError     = "application/vnd.crucible.error.v0+json"     // Media type for [RegistryError].
)

// Describes a namespace.
//
// This data structure is used when querying namespace metadata from the
// registry and listing available resources within that namespace. It provides
// a snapshot of the namespace's contents and basic information, but doesn't
// include detailed resource data. It should be used primarily as an index for
// discovering resources within a namespace.
//
// The corresponding media type is application/vnd.crucible.namespace.v0+json.
type NamespaceInfo struct {
	Namespace   string            `field:"namespace"`   // Namespace identifier.
	Description string            `field:"description"` // Human-readable description.
	Resources   []ResourceSummary `field:"resources"`   // Resources in the namespace.
}

// Describes a resource in a namespace listing.
//
// This data structure provides a brief overview of a resource within a
// namespace. It contains only discovery information. For full resource
// details including versions and channels, see [ResourceInfo].
type ResourceSummary struct {
	Name        string    `field:"name"`        // Resource name.
	Type        string    `field:"type"`        // Resource type (e.g., "template", "plugin").
	Description string    `field:"description"` // Human-readable description.
	Latest      string    `field:"latest"`      // Latest version string.
	UpdatedAt   time.Time `field:"updated_at"`  // Timestamp of the latest version publication.
}

// Describes a resource and its available versions.
//
// This data structure contains complete information about a resource, including
// all published versions and configured release channels. It is the primary
// structure for version resolution and resource discovery. The versions list is
// ordered by semantic version in descending order (newest first). The channels
// list contains all configured release tracks.
//
// The corresponding media type is application/vnd.crucible.resource.v0+json.
type ResourceInfo struct {
	Name        string        `field:"name"`        // Resource name.
	Type        string        `field:"type"`        // Resource type (e.g., "widget").
	Description string        `field:"description"` // Human-readable description.
	Versions    []VersionInfo `field:"versions"`    // Available versions.
	Channels    []ChannelInfo `field:"channels"`    // Available channels.
}

// Describes an available version of a resource.
//
// This data structure represents a single published version of a resource.
// The digest provides content verification, ensuring the downloaded archive
// matches what was published. The size field enables clients to display
// download progress and verify complete transfers.
type VersionInfo struct {
	Version   reference.Version `field:"version"`   // Semantic version.
	Digest    reference.Digest  `field:"digest"`    // Content digest.
	Published time.Time         `field:"published"` // Publication timestamp.
	Size      int64             `field:"size"`      // Archive size in bytes.
}

// Describes a release channel.
//
// A channel is a named pointer to a specific version. Channels provide release
// tracks such as "alpha" or "beta". Unlike versions, channels can be updated
// to point to different versions over time.
type ChannelInfo struct {
	VersionInfo
	Channel     string `field:"channel"`     // Channel name (e.g., "stable", "beta").
	Description string `field:"description"` // Human-readable description.
}

// Error returned by the registry.
//
// This data structure represents an error response from the registry API. The
// code field contains a machine-readable error identifier that clients can use
// for programmatic error handling. The message field provides a human-readable
// description of the error.
//
// The corresponding media type is application/vnd.crucible.error.v0+json.
type RegistryError struct {
	Code    string `field:"code"`    // Machine-readable error code.
	Message string `field:"message"` // Human-readable error description.
}

// Implements the error interface.
func (e *RegistryError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
