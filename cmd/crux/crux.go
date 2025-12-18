package main

import (
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/cli"
	"github.com/cruciblehq/crux/pkg/crex"
)

// The entry point for the Crux CLI application.
//
// It initializes logging, displays startup information, and executes the root
// command. If any error occurs during execution, it exits with a non-zero code.
func main() {

	// Set up buffering handler
	handler := crex.NewHandler()
	slog.SetDefault(slog.New(handler))

	// Log startup info (buffered until configured)
	slog.Debug("build", "version", internal.VersionString())

	slog.Debug("crux is running",
		"pid", os.Getpid(),
		"cwd", cwd(),
		"args", os.Args,
	)

	// Run (CLI configures and flushes the handler)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

// Returns the current working directory or "(unknown)".
func cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "(unknown)"
	}
	return cwd
}
