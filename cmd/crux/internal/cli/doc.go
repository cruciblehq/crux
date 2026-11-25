// Package cli parses arguments and dispatches subcommands for the crux CLI.
//
// Kong parses the arguments into [RootCmd] and calls the Run method of the
// selected subcommand. The subcommands build and package resources (build and
// pack), move resources between this computer and a registry (push and pull),
// save remote OCI images as local archives (import), manage the local resource
// cache (cache), manage the registries crux uses (registry), manage the local
// environment and its blueprint (local), and show version information (version).
//
// The flags on the root command apply to every subcommand. -C changes the
// working directory and is resolved to an absolute path before Run. -q and -v
// adjust verbosity, and -d enables debug output.
//
// Running the CLI:
//
//	if err := cli.Execute(ctx); err != nil {
//		log.Fatal(err)
//	}
package cli
