package cli

import (
	"context"
	"log/slog"

	"github.com/alecthomas/kong"
	"github.com/cruciblehq/crux/internal"
)

var RootCmd struct {

	// Global flags
	Verbose bool `short:"v" help:"Enable verbose output."`
	Debug   bool `short:"d" help:"Enable debug output."`

	// Subcommands
	// Scaffold ScaffoldCmd `cmd:"" aliases:"init,create,new" help:"Scaffold a Crucible resource."`
	Build BuildCmd `cmd:"" help:"Build and bundle Crucible resources."`
	// Server   ServerCmd   `cmd:"" help:"Manage the local development server."`

	// Version flag
	Version kong.VersionFlag `short:"V" help:"Show version information."`
}

// Parses arguments and runs the CLI
func Execute() error {

	ctx := kong.Parse(&RootCmd,
		kong.Name("crux"),
		kong.Description("Crucible's build tool.\n\nEnables scaffolding, testing, building, packaging, and distributing of Crucible resources."),
		kong.UsageOnError(),
		kong.Vars{
			"version": internal.VersionString(),
		},
	)

	if err := ctx.Run(context.Background()); err != nil {
		slog.Error("command failed", "error", err)
		return err
	}

	return nil
}
