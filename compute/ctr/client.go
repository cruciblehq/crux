package ctr

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/crex"
)

// Containerd namespace used for all Crucible containers.
const crucibleNamespace = "crucible"

// OCI runtime used for Linux build and run containers.
const runcV2Runtime = "io.containerd.runc.v2"

// containerd-backed [provider.ContainerRuntime].
type runtime struct {
	client *containerd.Client
}

// Connects to the containerd socket at socketPath.
//
// Returns a [provider.ContainerRuntime] backed by the daemon at socketPath.
// The caller must call Close when done.
func New(socketPath string) (provider.ContainerRuntime, error) {
	c, err := containerd.New(socketPath)
	if err != nil {
		return nil, crex.Wrap(ErrConnect, err)
	}
	return &runtime{client: c}, nil
}

// Loads an OCI image archive into the containerd image store.
//
// Returns the image reference for use with [Run].
func (r *runtime) Import(ctx context.Context, reader io.Reader) (string, error) {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)
	imgs, err := r.client.Import(ctx, reader)
	if err != nil {
		return "", crex.Wrap(ErrImport, err)
	}
	if len(imgs) == 0 {
		return "", crex.Wrap(ErrImport, ErrNoImages)
	}
	return imgs[0].Name, nil
}

// Creates, starts, and waits for a container, then removes it.
//
// The spec from [provider.ExecConfig.Security] is merged with the image's own
// process configuration: image env, entrypoint, workdir, and user provide
// baseline values; security and resource policy from the subsystems overrides
// those fields. Infrastructure fields (rootfs path, cgroup path) are managed
// by containerd.
//
// Returns the container's exit code, or -1 and an error if the container
// could not be started or the context was cancelled.
func (r *runtime) Run(ctx context.Context, cfg *provider.ExecConfig) (int, error) {
	ctx = namespaces.WithNamespace(ctx, crucibleNamespace)

	img, err := r.client.GetImage(ctx, cfg.Image)
	if err != nil {
		return -1, crex.Wrapf(ErrRun, "get image %q: %w", cfg.Image, err)
	}

	snapshotID := cfg.ID + "-rootfs"

	platform, err := imagePlatform(ctx, img, r.client.ContentStore())
	if err != nil {
		return -1, crex.Wrapf(ErrRun, "get image platform: %w", err)
	}

	container, err := r.client.NewContainer(ctx, cfg.ID,
		containerd.WithRuntime(runcV2Runtime, nil),
		containerd.WithImage(img),
		containerd.WithNewSnapshot(snapshotID, img),
		containerd.WithNewSpec(
			oci.WithDefaultSpecForPlatform(platform),
			oci.WithImageConfig(img),
			withSecurityPolicy(cfg.Security),
			withProcessConfig(cfg),
		),
	)
	if err != nil {
		return -1, crex.Wrapf(ErrRun, "create container: %w", err)
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	// FIFOs do not work across the macOS/Lima VM kernel boundary via virtiofs.
	// Use a log file in the virtiofs-accessible cache dir instead.
	logDir, _ := os.UserCacheDir()
	logDir = filepath.Join(logDir, "crux", "logs")
	os.MkdirAll(logDir, 0o700)
	logPath := filepath.Join(logDir, cfg.ID+".log")
	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); err == nil {
		f.Close()
	}
	defer os.Remove(logPath)

	task, err := container.NewTask(ctx, cio.LogFile(logPath))
	if err != nil {
		return -1, crex.Wrapf(ErrRun, "create task: %w", err)
	}

	exitCh, err := task.Wait(ctx)
	if err != nil {
		task.Delete(ctx)
		return -1, crex.Wrapf(ErrRun, "wait: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		return -1, crex.Wrapf(ErrRun, "start task: %w", err)
	}

	var exitCode int
	select {
	case status := <-exitCh:
		task.Delete(ctx)
		if err := status.Error(); err != nil {
			return -1, crex.Wrapf(ErrRun, "exit status: %w", err)
		}
		exitCode = int(status.ExitCode())
	case <-ctx.Done():
		task.Kill(ctx, syscall.SIGTERM)
		task.Delete(ctx)
		return -1, ctx.Err()
	}

	// Forward log output to the process stdio.
	if data, err := os.ReadFile(logPath); err == nil && len(data) > 0 {
		os.Stdout.Write(data)
	}

	return exitCode, nil
}

// Closes the connection to the containerd daemon.
func (r *runtime) Close() error {
	return r.client.Close()
}

// Returns an [oci.SpecOpts] that overlays the security and resource policy
// from s onto the spec produced by containerd.
//
// Applied fields: capabilities, rlimits, no-new-privileges, seccomp, cgroup
// resources, and Linux namespaces. Infrastructure fields (Root.Path,
// CgroupsPath) are left for containerd to manage.
func withSecurityPolicy(s *specs.Spec) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, spec *specs.Spec) error {
		if s == nil {
			return nil
		}
		if s.Process != nil {
			spec.Process.Capabilities = s.Process.Capabilities
			spec.Process.Rlimits = s.Process.Rlimits
			spec.Process.NoNewPrivileges = s.Process.NoNewPrivileges
		}
		if s.Linux != nil {
			spec.Linux.Seccomp = s.Linux.Seccomp
			spec.Linux.Resources = s.Linux.Resources
			if len(s.Linux.Namespaces) > 0 {
				spec.Linux.Namespaces = s.Linux.Namespaces
			}
		}
		return nil
	}
}

// Returns an [oci.SpecOpts] that applies the process configuration from cfg.
//
// Args overrides the image entrypoint when non-empty. Env is merged onto the
// image environment with caller values winning on key conflicts. WorkingDir
// overrides the image default when non-empty.
func withProcessConfig(cfg *provider.ExecConfig) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, spec *specs.Spec) error {
		if len(cfg.Args) > 0 {
			spec.Process.Args = cfg.Args
		}
		if len(cfg.Env) > 0 {
			spec.Process.Env = mergeEnv(spec.Process.Env, cfg.Env)
		}
		if cfg.WorkingDir != "" {
			spec.Process.Cwd = cfg.WorkingDir
		}
		return nil
	}
}

// Merges base environment variables with overrides, caller values winning.
//
// Each entry is "KEY=value". When the same key appears in both slices the
// override entry is kept and the base entry is dropped. Entries without an
// '=' separator are kept as-is from base.
func mergeEnv(base, overrides []string) []string {
	// Index override keys for O(1) lookup during base filtering.
	overrideKeys := make(map[string]struct{}, len(overrides))
	for _, e := range overrides {
		if idx := strings.IndexByte(e, '='); idx > 0 {
			overrideKeys[e[:idx]] = struct{}{}
		}
	}

	out := make([]string, 0, len(base)+len(overrides))
	for _, e := range base {
		if idx := strings.IndexByte(e, '='); idx > 0 {
			if _, shadowed := overrideKeys[e[:idx]]; shadowed {
				continue
			}
		}
		out = append(out, e)
	}
	return append(out, overrides...)
}

// Returns the OCI platform string (e.g. "linux/amd64") for the given image by
// reading the image config. Falls back to "linux/amd64" when the config does
// not specify OS or architecture.
func imagePlatform(ctx context.Context, img containerd.Image, cs content.Store) (string, error) {
	configDesc, err := img.Config(ctx)
	if err != nil {
		return "", err
	}
	data, err := content.ReadBlob(ctx, cs, configDesc)
	if err != nil {
		return "", err
	}
	var cfg ocispec.Image
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	os := cfg.OS
	if os == "" {
		os = "linux"
	}
	arch := cfg.Architecture
	if arch == "" {
		arch = "amd64"
	}
	return os + "/" + arch, nil
}
