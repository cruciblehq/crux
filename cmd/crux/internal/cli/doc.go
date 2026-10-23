// Package cli parses arguments and dispatches subcommands for the crux CLI.
//
// Provides usage information and examples for each subcommand. Each subcommand
// has a Run method that is dispatched by Kong after parsing. This includes
// commands for building and packaging resources (build and pack), running
// resources locally (start, stop, restart, reset, destroy, exec, and status),
// moving resources between machines and registries (push, pull, and import),
// inspecting local state (cache, host, and version), and managing the host
// environment. Each command resolves its own flags and arguments to absolute
// values before Run.
//
// Global flags (on the root command) apply globally. -C changes the working
// directory, -q and -v adjust verbosity, and -d enables debug output.
//
// Wiring the parser and dispatching a command:
//
//	var root cli.Root
//	ctx := kong.Parse(&root)
//	if err := ctx.Run(); err != nil {
//		log.Fatal(err)
//	}
package cli
