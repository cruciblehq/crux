package main

import (
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/cli"
	"github.com/cruciblehq/crex"
)

// The entry point for the Crux CLI application.
//
// It initializes logging, displays startup information, and executes the root
// command. If any error occurs during execution, it exits with a non-zero code.
func main() {
	slog.SetDefault(logger())

	slog.Debug("build",
		"version", internal.VersionString(),
	)

	slog.Debug("crux is running",
		"pid", os.Getpid(),
		"cwd", cwd(),
		"args", os.Args,
	)

	if err := cli.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// Creates a buffered logger for CLI operation.
//
// The initial log level is derived from build-time linker flags. The CLI
// reconfigures and flushes the handler after parsing flags.
func logger() *slog.Logger {
	handler := crex.NewHandler()
	handler.SetLevel(logLevel())
	return slog.New(handler.WithGroup(internal.Name))
}

// Returns the log level derived from build-time linker flags.
func logLevel() slog.Level {
	if internal.IsDebug() {
		return slog.LevelDebug
	}
	if internal.IsQuiet() {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}

// Returns the current working directory or "(unknown)".
func cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "(unknown)"
	}
	return cwd
}
