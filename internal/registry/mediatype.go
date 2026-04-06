package registry

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
