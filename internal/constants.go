package internal

const (

	// Identifier for the cruxd instance, also used as the slog group name.
	InstanceName = "crux"

	// Crucible Hub registry URL.
	RegistryURL = "http://hub.cruciblehq.xyz:8080"

	// Default namespace for resources in the registry.
	DefaultNamespace = "official"

	// Cruxd version to provision.
	//
	// This version is embedded at build time and determines which cruxd release
	// is downloaded during provisioning. It should be bumped whenever a new
	// cruxd release is adopted.
	CruxdVersion = "0.3.1"
)
