package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/cli"
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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Restore default signal behaviour after the first cancellation so a
	// second CTRL-C terminates the process immediately.
	go func() {
		<-ctx.Done()
		stop()
	}()

	if err := cli.Execute(ctx); err != nil {
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
