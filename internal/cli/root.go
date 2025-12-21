package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/pkg/crex"
)

var RootCmd struct {

	// Global flags
	Verbose bool `short:"v" help:"Enable verbose output."`
	Debug   bool `short:"d" help:"Enable debug output."`

	// Subcommands
	// Scaffold ScaffoldCmd `cmd:"" aliases:"init,create,new" help:"Scaffold a Crucible resource."`
	Build   BuildCmd   `cmd:"" help:"Build and bundle Crucible resources."`
	Version VersionCmd `cmd:"" help:"Show version information."`
	// Server   ServerCmd   `cmd:"" help:"Manage the local development server."`
}

// Parses arguments and runs the CLI
func Execute() error {

	ctx := context.Background()

	kongCtx := kong.Parse(&RootCmd,
		kong.Name("crux"),
		kong.Description("Crucible's build tool.\n\nEnables scaffolding, testing, building, packaging, and distributing of Crucible resources."),
		kong.UsageOnError(),
		kong.Vars{
			"version": internal.VersionString(),
		},
		kong.BindTo(ctx, (*context.Context)(nil)),
	)

	// Configure logger based on parsed flags
	configureLogger()

	if err := kongCtx.Run(); err != nil {
		slog.Error("command failed", "error", err)
		return err
	}

	return nil
}

func configureLogger() {

	// Get the handler from the default logger
	handler, ok := slog.Default().Handler().(crex.Handler)
	if !ok {
		return // Not a crex.Handler, nothing to configure
	}

	// Formatter
	formatter := crex.NewPrettyFormatter(isatty(os.Stderr))
	formatter.SetVerbose(RootCmd.Verbose)
	handler.SetFormatter(formatter)

	// Set log level based on flags
	if RootCmd.Debug {
		handler.SetLevel(slog.LevelDebug)
	} else if RootCmd.Verbose {
		handler.SetLevel(slog.LevelInfo)
	} else {
		handler.SetLevel(slog.LevelWarn)
	}

	// Output stream
	handler.SetStream(os.Stderr)
	handler.Flush()
}

// Whehther the given file is an interactive terminal.
func isatty(f *os.File) bool {
	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
