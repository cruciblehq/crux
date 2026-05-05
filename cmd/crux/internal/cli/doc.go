// Package cli parses arguments and dispatches subcommands for the crux CLI.
//
// Each subcommand is a separate type with a Run method that Kong dispatches
// after parsing. The available commands cover building and packaging
// resources through build and pack, running them locally through start,
// stop, restart, reset, destroy, exec, and status, moving them between
// machines and registries through push, pull, and import, inspecting local
// state through cache, host, and version, and managing the surrounding
// Crucible host environment. Each command type owns the flags and arguments
// specific to its operation and resolves them to absolute values before its
// Run method begins, so the body of every command operates on fully validated
// input.
//
// Global flags on the root command reconfigure the logger before any
// subcommand executes, so logging behaves consistently regardless of which
// command runs. The -C flag changes the working directory before resolving
// any relative paths, -q and -v adjust verbosity, and -d enables debug output.
//
// Wiring the parser and dispatching a command:
//
//	var root cli.Root
//	ctx := kong.Parse(&root)
//	if err := ctx.Run(); err != nil {
//		log.Fatal(err)
//	}
package cli
