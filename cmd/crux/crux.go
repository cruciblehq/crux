package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/cmd/crux/internal/cli"
	"github.com/cruciblehq/crux/crex"
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

	ctx, stop := setUpSignalHandler()
	defer stop()

	if err := cli.Execute(ctx); err != nil {
		crex.LogError(slog.Default(), err)
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

// Sets up a signal handler.
//
// The handler cancels a context on the first SIGINT or SIGTERM received. After
// the first signal, default behaviour is restored so a second signal terminates
// the process immediately.
func setUpSignalHandler() (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	// Restore default behaviour on second signal.
	go func() {
		<-ctx.Done()
		stop()
	}()

	return ctx, stop
}

// Returns the log level derived from build-time linker flags.
//
// Debug builds default to debug level, quiet builds default to warn level, and
// all others default to info level.
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
