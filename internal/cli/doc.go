// Parses arguments and dispatches subcommands for the crux CLI.
//
// Each subcommand is a separate type with a Run method that Kong dispatches
// after parsing. The available commands are:
//
//	build       Build and bundle Crucible resources.
//	pack        Package a built resource for distribution.
//	push        Push a resource package to the Hub registry.
//	pull        Pull a resource from the Hub registry to local cache.
//	plan        Generate a deployment plan from a blueprint.
//	provider    Manage cloud provider configurations.
//	cache       Manage the local resource cache.
//	runtime     Manage the container runtime environment.
//	image       Manage OCI images in the runtime.
//	container   Manage containers in the runtime.
//	version     Show version information.
//
// Global flags (-C, -q, -v, -d) live on the root command and reconfigure the
// logger before any subcommand executes.
package cli
