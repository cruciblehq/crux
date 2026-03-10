package internal

const (

	// Application name, used as the slog group name.
	Name = "crux"

	// Default cruxd instance name used when no instance is specified.
	DefaultInstanceName = "local"

	// Default Crucible Hub registry URL.
	DefaultRegistryURL = "http://hub.cruciblehq.xyz:8080"

	// Default namespace for resources in the registry.
	DefaultNamespace = "official"

	// Default machine image reference for the runtime VM (Darwin only).
	//
	// This reference is embedded at build time and determines which machine
	// resource is pulled from the registry during provisioning. The machine
	// image includes cruxd, containerd, and all required services pre-
	// configured. Bump this when adopting a new machine image release.
	DefaultMachineImage = "crucible/machine 0.1.0"
)
