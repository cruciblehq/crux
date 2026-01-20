package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/pkg/crex"
)

// Represents the root command for the Crux CLI.
var RootCmd struct {

	// Global flags
	Verbose bool `short:"v" help:"Enable verbose output."`
	Debug   bool `short:"d" help:"Enable debug output."`

	// Subcommands
	// Scaffold ScaffoldCmd `cmd:"" aliases:"init,create,new" help:"Scaffold a Crucible resource."`
	Build    BuildCmd    `cmd:"" help:"Build and bundle Crucible resources."`
	Pack     PackCmd     `cmd:"" help:"Package a built resource for distribution."`
	Push     PushCmd     `cmd:"" help:"Push a resource package to the Hub registry."`
	Plan     PlanCmd     `cmd:"" help:"Generate a deployment plan from a blueprint."`
	Deploy   DeployCmd   `cmd:"" help:"Deploy a blueprint using a plan."`
	Provider ProviderCmd `cmd:"" help:"Manage cloud provider configurations."`
	Version  VersionCmd  `cmd:"" help:"Show version information."`
	// Server   ServerCmd   `cmd:"" help:"Manage the local development server."`
}

// Parses arguments and runs the CLI.
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

	configureLogger()

	if err := kongCtx.Run(); err != nil {
		return err
	}

	return nil
}

// Configures the global logger based on CLI flags.
func configureLogger() {

	// Get the handler from the default logger
	handler, ok := slog.Default().Handler().(crex.Handler)
	if !ok {
		return // Not a crex.Handler, nothing to configure
	}

	// Configure formatter
	formatter := crex.NewPrettyFormatter(isatty(os.Stderr))
	formatter.SetVerbose(RootCmd.Verbose)

	// Configure handler
	if RootCmd.Debug {
		handler.SetLevel(slog.LevelDebug)
	} else if RootCmd.Verbose {
		handler.SetLevel(slog.LevelInfo)
	} else {
		handler.SetLevel(slog.LevelWarn)
	}

	// Commit
	handler.SetFormatter(formatter)
	handler.SetStream(os.Stderr)
	handler.Flush()
}

// Whether the given file is an interactive terminal.
func isatty(f *os.File) bool {
	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
