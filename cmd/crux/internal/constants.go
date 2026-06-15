package internal

// Application name, used as the slog group name.
const Name = "crux"

// Defaults for compute host, registry, and namespace.
const (

	// Default compute host name.
	DefaultInstanceName = "local"

	// Default Crucible Hub registry URL.
	DefaultRegistryURL = "http://hub.cruciblehq.xyz:8080"

	// Default namespace for resources in the registry.
	DefaultNamespace = "official"
)
