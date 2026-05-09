package cli

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/crex"
)

// Represents the root command for the Crux CLI.
var RootCmd struct {
	Context string `short:"C" help:"Run as if crux was started in the given directory." default:"."`
	Quiet   bool   `short:"q" help:"Suppress informational output."`
	Verbose bool   `short:"v" help:"Enable verbose output."`
	Debug   bool   `short:"d" help:"Enable debug output."`

	Build   BuildCmd   `cmd:"" help:"Build and bundle Crucible resources."`
	Pack    PackCmd    `cmd:"" help:"Package a built resource for distribution."`
	Push    PushCmd    `cmd:"" help:"Push a resource package to the Hub registry."`
	Pull    PullCmd    `cmd:"" help:"Pull a resource from the Hub registry to local cache."`
	Import  ImportCmd  `cmd:"" help:"Pull a remote OCI image and save it as a local archive."`
	Cache   CacheCmd   `cmd:"" help:"Manage the local resource cache."`
	Host    HostCmd    `cmd:"" help:"Manage the Crucible host environment."`
	Version VersionCmd `cmd:"" help:"Show version information."`
}

// Parses arguments and runs the CLI.
func Execute(ctx context.Context) error {

	kongCtx := kong.Parse(&RootCmd,
		kong.Name("crux"),
		kong.Description("Crucible resource manager.\n\nBuild, run, and distribute Crucible resources."),
		kong.UsageOnError(),
		kong.Vars{
			"version": internal.VersionString(),
		},
		kong.BindTo(ctx, (*context.Context)(nil)),
	)

	configureLogger()

	// Resolve -C to an absolute path so all downstream consumers get a fully
	// qualified path.
	if abs, err := filepath.Abs(RootCmd.Context); err == nil {
		RootCmd.Context = abs
	}

	return kongCtx.Run()
}

// Configures the global logger based on CLI flags.
func configureLogger() {
	handler, ok := slog.Default().Handler().(crex.Handler)
	if !ok {
		return // Not a crex.Handler, nothing to configure
	}

	// Configure formatter
	formatter := crex.NewPrettyFormatter(isatty(os.Stderr))
	formatter.Verbose = RootCmd.Verbose

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
