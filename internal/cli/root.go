package cli

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
)

// Represents the root command for the Crux CLI.
var RootCmd struct {

	// Global flags
	Context string `short:"C" help:"Run as if crux was started in the given directory." default:"."`
	Quiet   bool   `short:"q" help:"Suppress informational output."`
	Verbose bool   `short:"v" help:"Enable verbose output."`
	Debug   bool   `short:"d" help:"Enable debug output."`

	// Subcommands
	Build   BuildCmd   `cmd:"" help:"Build and bundle Crucible resources."`
	Pack    PackCmd    `cmd:"" help:"Package a built resource for distribution."`
	Start   StartCmd   `cmd:"" help:"Start a resource."`
	Stop    StopCmd    `cmd:"" help:"Stop a running resource."`
	Restart RestartCmd `cmd:"" help:"Restart a resource."`
	Reset   ResetCmd   `cmd:"" help:"Destroy and recreate a resource."`
	Destroy DestroyCmd `cmd:"" help:"Remove a resource and its runtime state."`
	Exec    ExecCmd    `cmd:"" help:"Execute a command inside a running resource."`
	Status  StatusCmd  `cmd:"" help:"Show the state of a resource."`
	Push    PushCmd    `cmd:"" help:"Push a resource package to the Hub registry."`
	Pull    PullCmd    `cmd:"" help:"Pull a resource from the Hub registry to local cache."`
	Cache   CacheCmd   `cmd:"" help:"Manage the local resource cache."`
	Host    HostCmd    `cmd:"" help:"Manage the Crucible host environment."`
	Version VersionCmd `cmd:"" help:"Show version information."`
	// Scaffold ScaffoldCmd `cmd:"" aliases:"init,create,new" help:"Scaffold a Crucible resource."`
	// Server   ServerCmd   `cmd:"" help:"Manage the local development server."`
}

// Parses arguments and runs the CLI.
func Execute(ctx context.Context) error {

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

	// Resolve -C to an absolute path so all downstream consumers
	// (including cruxd inside the VM) get a fully qualified path.
	if abs, err := filepath.Abs(RootCmd.Context); err == nil {
		RootCmd.Context = abs
	}

	if err := kongCtx.Run(); err != nil {
		return err
	}

	return nil
}

// Configures the global logger based on CLI flags.
func configureLogger() {
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
	} else if RootCmd.Quiet {
		handler.SetLevel(slog.LevelWarn)
	} else {
		handler.SetLevel(slog.LevelInfo)
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
