// Parses arguments and dispatches subcommands for the crux CLI.
//
// Each subcommand is a separate type with a Run method that Kong dispatches
// after parsing. The available commands are:
//
//	build       Build and bundle Crucible resources.
//	pack        Package a built resource for distribution.
//	start       Start a resource.
//	stop        Stop a running resource.
//	destroy     Remove a resource and its runtime state.
//	exec        Execute a command inside a running resource.
//	status      Show the state of a resource.
//	push        Push a resource package to the Hub registry.
//	pull        Pull a resource from the Hub registry to local cache.
//	cache       Manage the local resource cache.
//	runtime     Manage the container runtime environment.
//	version     Show version information.
//
// Global flags (-C, -q, -v, -d) live on the root command and reconfigure the
// logger before any subcommand executes.
package cli
