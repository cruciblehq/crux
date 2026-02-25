package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/manifest"
)

// Runner for Crucible runtimes.
//
// Extracts the runtime configuration from the manifest and delegates to the
// shared recipe pipeline. Only the Build and Pack operations are supported.
type RuntimeRunner struct {
	recipeBuilder
}

// Returns a [RuntimeRunner] wired to the given daemon client.
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewRuntimeRunner(client *daemon.Client, registry, defaultNamespace, workdir string) *RuntimeRunner {
	return &RuntimeRunner{
		recipeBuilder: recipeBuilder{
			client:           client,
			registry:         registry,
			defaultNamespace: defaultNamespace,
			workdir:          workdir,
		},
	}
}

// Builds a Crucible runtime resource based on the provided manifest.
//
// The runtime configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (rr *RuntimeRunner) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, ok := m.Config.(*manifest.Runtime)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	return rr.build(ctx, m, &cfg.Recipe, output, nil)
}

func (rr *RuntimeRunner) Start(_ context.Context, _ manifest.Manifest, _ string) error {
	return ErrUnsupported
}

func (rr *RuntimeRunner) Stop(_ context.Context, _ manifest.Manifest) error {
	return ErrUnsupported
}

func (rr *RuntimeRunner) Destroy(_ context.Context, _ manifest.Manifest) error {
	return ErrUnsupported
}

func (rr *RuntimeRunner) Exec(_ context.Context, _ manifest.Manifest, _ []string) (*ExecResult, error) {
	return nil, ErrUnsupported
}

func (rr *RuntimeRunner) Status(_ context.Context, _ manifest.Manifest) (*StatusResult, error) {
	return nil, ErrUnsupported
}

// Packages the runtime's build output into a distributable archive.
//
// The dist directory must contain image.tar.
func (rr *RuntimeRunner) Pack(ctx context.Context, m manifest.Manifest, manifestPath, dist, output string) (*PackResult, error) {
	return pack(ctx, m, manifestPath, dist, output)
}

// Uploads a runtime package archive to the Hub registry.
//
// packagePath must point to an archive created by [RuntimeRunner.Pack].
func (rr *RuntimeRunner) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, m, packagePath, rr.registry, rr.defaultNamespace)
}
