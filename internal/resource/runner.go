package resource

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/manifest"
)

// Manages the lifecycle of a single Crucible resource type.
//
// Each resource type provides its own Runner implementation. A Runner is
// obtained via [Resolve]. Not every resource type supports every operation.
// For example, widgets do not support Exec. Unsupported operations return
// [ErrUnsupported], allowing callers to handle them gracefully.
type Runner interface {

	// Compiles the resource according to its manifest and writes the resulting
	// artifacts to the output directory.
	Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error)

	// Imports a built artifact and starts the resource.
	//
	// path points to the build output required by the resource type (e.g. an
	// OCI image tar for services and runtimes).
	Start(ctx context.Context, m manifest.Manifest, path string) error

	// Gracefully stops a running resource without removing its state.
	//
	// The resource can be restarted with [Runner.Start].
	Stop(ctx context.Context, m manifest.Manifest) error

	// Stops the resource and starts it again, preserving filesystem state.
	//
	// path points to the build output required by the resource type.
	Restart(ctx context.Context, m manifest.Manifest, path string) error

	// Destroys the resource and starts it fresh from the image.
	//
	// path points to the build output required by the resource type.
	Reset(ctx context.Context, m manifest.Manifest, path string) error

	// Stops the resource if running and removes all of its runtime state.
	Destroy(ctx context.Context, m manifest.Manifest) error

	// Runs a command inside a running resource and returns the result.
	//
	// The command is executed in the resource's default environment.
	Exec(ctx context.Context, m manifest.Manifest, command []string) (*ExecResult, error)

	// Returns the current state of a resource (e.g. running, stopped).
	Status(ctx context.Context, m manifest.Manifest) (*StatusResult, error)

	// Packages built artifacts into a distributable archive.
	//
	// The archive layout is type-specific: each resource type decides which
	// files are included and how they are structured. The manifest file at
	// manifestPath is included in the archive. The output extension must be .tar.zst.
	Pack(ctx context.Context, m manifest.Manifest, manifestPath, dist, output string) (*PackResult, error)

	// Pushes a packaged resource archive to the registry.
	//
	// packagePath must point to an archive created by [Runner.Pack]. The
	// target registry is determined by the resource name in the manifest,
	// falling back to the DefaultRegistry provided in [Options].
	Push(ctx context.Context, m manifest.Manifest, packagePath string) error
}

// Configures a [Runner] obtained through [Resolve].
//
// DefaultRegistry and DefaultNamespace are required for operations that
// interact with a registry (Build, Push). They can be omitted for lifecycle
// operations (Start, Stop, Destroy, Exec, Status).
type Options struct {
	DefaultRegistry  string // Fallback registry URL when the resource name does not include one.
	DefaultNamespace string // Fallback namespace when the resource name does not include one.
}

// Holds the output of a successful [Runner.Build] call.
type BuildResult struct {
	Output   string             // Directory where the build artifacts were written.
	Manifest *manifest.Manifest // The fully resolved manifest used for the build.
}

// Holds the output of a [Runner.Exec] call.
type ExecResult struct {
	ExitCode int    // Process exit code.
	Stdout   string // Captured standard output.
	Stderr   string // Captured standard error.
}

// Holds the output of a [Runner.Status] call.
type StatusResult struct {
	Status string // Current state of the resource (e.g. "running", "stopped").
}

// Reads the manifest at the given path and returns the appropriate [Runner]
// for the resource type declared in it.
//
// The returned manifest is fully decoded and ready to be passed to Runner
// methods. Options are accepted variadically for convenience; only the
// first value is used.
func Resolve(manifestPath string, opts ...Options) (*manifest.Manifest, Runner, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, nil, crex.Wrap(ErrRunner, err)
	}

	man, err := manifest.Decode(data)
	if err != nil {
		return nil, nil, crex.Wrap(ErrRunner, err)
	}

	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	workdir := filepath.Dir(manifestPath)

	var r Runner
	switch man.Resource.Type {
	case manifest.TypeRuntime:
		client := daemon.NewClient()
		r = NewRuntimeRunner(client, o.DefaultRegistry, o.DefaultNamespace, workdir)
	case manifest.TypeService:
		client := daemon.NewClient()
		r = NewServiceRunner(client, o.DefaultRegistry, o.DefaultNamespace, workdir)
	case manifest.TypeWidget:
		r = NewWidgetRunner(o.DefaultRegistry, o.DefaultNamespace)
	default:
		r = &runnerStub{resourceType: string(man.Resource.Type)}
	}

	return man, r, nil
}
