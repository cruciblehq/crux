package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/daemon"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/protocol"
)

// [Runner] for Crucible services.
//
// Supports the full Runner lifecycle: Build, Start, Stop, Destroy, Exec,
// Status, Pack, and Push. Embeds [recipeBuilder] for the build pipeline.
type ServiceRunner struct {
	recipeBuilder
}

// Returns a [ServiceRunner] wired to the given daemon client.
//
// workdir is the directory containing the manifest and is used as the root
// for resolving copy sources during builds.
func NewServiceRunner(client *daemon.Client, registry, defaultNamespace, workdir string) *ServiceRunner {
	return &ServiceRunner{
		recipeBuilder: recipeBuilder{
			client:           client,
			registry:         registry,
			defaultNamespace: defaultNamespace,
			workdir:          workdir,
		},
	}
}

// Builds a Crucible service resource based on the provided manifest.
//
// The service configuration is extracted and the shared recipe pipeline
// handles the build process. The built artifacts are placed in the directory
// specified by the output parameter.
func (r *ServiceRunner) Build(ctx context.Context, m manifest.Manifest, output string) (*BuildResult, error) {
	cfg, ok := m.Config.(*manifest.Service)
	if !ok {
		return nil, crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	return r.build(ctx, m, &cfg.Recipe, output, cfg.Entrypoint)
}

// Ensures the service is running.
//
// The operation is idempotent: if the container is already running it is
// left untouched; if it is stopped a new process is started on the
// existing snapshot; if it does not exist the image is imported and a
// fresh container is created.
func (r *ServiceRunner) Start(ctx context.Context, m manifest.Manifest, path string) error {
	id := m.Resource.Name

	result, err := r.client.ContainerStatus(ctx, &protocol.ContainerStatusRequest{ID: id})
	if err != nil {
		return crex.Wrap(ErrRunner, err)
	}

	switch result.Status {
	case protocol.ContainerRunning:
		return nil

	case protocol.ContainerStopped:
		if err := r.client.ImageStart(ctx, &protocol.ImageStartRequest{
			Ref:     m.Resource.Name,
			Version: m.Resource.Version,
		}); err != nil {
			return crex.Wrap(ErrRunner, err)
		}
		return nil

	default:
		if err := r.client.ImageImport(ctx, &protocol.ImageImportRequest{
			Ref:     m.Resource.Name,
			Version: m.Resource.Version,
			Path:    path,
		}); err != nil {
			return crex.Wrap(ErrRunner, err)
		}

		if err := r.client.ImageStart(ctx, &protocol.ImageStartRequest{
			Ref:     m.Resource.Name,
			Version: m.Resource.Version,
		}); err != nil {
			return crex.Wrap(ErrRunner, err)
		}
		return nil
	}
}

// Stops the service and starts it again, preserving the container snapshot.
func (r *ServiceRunner) Restart(ctx context.Context, m manifest.Manifest, path string) error {
	r.client.ContainerStop(ctx, &protocol.ContainerStopRequest{ID: m.Resource.Name})
	return r.Start(ctx, m, path)
}

// Destroys the service container and starts fresh from the image.
func (r *ServiceRunner) Reset(ctx context.Context, m manifest.Manifest, path string) error {
	r.client.ContainerDestroy(ctx, &protocol.ContainerDestroyRequest{ID: m.Resource.Name})
	return r.Start(ctx, m, path)
}

// Sends a graceful stop signal to the service container.
//
// Returns an error if the container is not running or the daemon is unreachable.
func (r *ServiceRunner) Stop(ctx context.Context, m manifest.Manifest) error {
	if err := r.client.ContainerStop(ctx, &protocol.ContainerStopRequest{
		ID: m.Resource.Name,
	}); err != nil {
		return crex.Wrap(ErrRunner, err)
	}
	return nil
}

// Stops and removes the service and its associated state.
func (r *ServiceRunner) Destroy(ctx context.Context, m manifest.Manifest) error {
	if err := r.client.ContainerDestroy(ctx, &protocol.ContainerDestroyRequest{
		ID: m.Resource.Name,
	}); err != nil {
		return crex.Wrap(ErrRunner, err)
	}
	return nil
}

// Runs a command inside the service's running container and returns the
// combined stdout, stderr, and exit code.
//
// The service must be running.
func (r *ServiceRunner) Exec(ctx context.Context, m manifest.Manifest, command []string) (*ExecResult, error) {
	result, err := r.client.ContainerExec(ctx, &protocol.ContainerExecRequest{
		ID:      m.Resource.Name,
		Command: command,
	})
	if err != nil {
		return nil, crex.Wrap(ErrRunner, err)
	}

	return &ExecResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
	}, nil
}

// Queries the daemon for the service container's current state (e.g.
// running, stopped, or not found).
func (r *ServiceRunner) Status(ctx context.Context, m manifest.Manifest) (*StatusResult, error) {
	result, err := r.client.ContainerStatus(ctx, &protocol.ContainerStatusRequest{
		ID: m.Resource.Name,
	})
	if err != nil {
		return nil, crex.Wrap(ErrRunner, err)
	}

	return &StatusResult{
		Status: string(result.Status),
	}, nil
}

// Packages the service's build output into a distributable archive.
//
// The dist directory must contain image.tar.
func (r *ServiceRunner) Pack(ctx context.Context, m manifest.Manifest, manifestPath, dist, output string) (*PackResult, error) {
	return pack(ctx, m, manifestPath, dist, output)
}

// Uploads a service package archive to the Hub registry.
//
// packagePath must point to an archive created by [ServiceRunner.Pack].
func (r *ServiceRunner) Push(ctx context.Context, m manifest.Manifest, packagePath string) error {
	return push(ctx, m, packagePath, r.registry, r.defaultNamespace)
}
