package build

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/manifest"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progresswriter"
)

const (

	// Dockerfile syntax identifier for Buildkit
	DockerfileFrontend = "dockerfile.v0"

	// File name for the exported service image tarball
	ServiceImageFileName = "image.tar"
)

// Defines the target platforms for Crucible service deployments. Services are
// built as multi-platform images to support the Crucible infrastructure.
var RuntimePlatforms = []string{
	"linux/amd64",
	"linux/arm64",
}

// Builder for Crucible services.
type ServiceBuilder struct{}

// Creates a new instance of [ServiceBuilder].
func NewServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{}
}

// Builds a Crucible service based on the provided manifest.
//
// It connects to Buildkit, prepares the build options based on the service's
// build configuration, invokes the build process, and streams the build
// progress to the console. The resulting container image is exported as an OCI
// tarball in the 'dist' directory.
func (sb *ServiceBuilder) Build(ctx context.Context, m manifest.Manifest) error {

	// Correct manifest type?
	service, ok := m.Config.(*manifest.Service)
	if !ok {
		return crex.ProgrammingError("build failed", "an internal configuration type mismatch occurred").
			Fallback("Please report this issue to the Crucible team.").
			Err()
	}

	// Connect to Buildkit
	socketPath := buildkitSocketPath()
	socketAddr := "unix://" + socketPath

	socketClient, err := client.New(ctx, socketAddr)
	if err != nil {
		return crex.UserError("build failed", "buildkitd is not running").
			Fallback("Install BuildKit with 'brew install buildkit' and start it with 'buildkitd --rootless &'").
			Err()
	}
	defer socketClient.Close()

	// Prepare build options
	solveOpt := client.SolveOpt{
		Frontend: DockerfileFrontend,
		FrontendAttrs: map[string]string{
			"filename": service.Build.Main, // Defaults to "Dockerfile" if empty
			"platform": strings.Join(RuntimePlatforms, ","),
		},
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterOCI,
				Output: func(map[string]string) (io.WriteCloser, error) {
					return os.Create(filepath.Join(Dist, ServiceImageFileName))
				},
			},
		},
	}

	// Add build args
	for k, v := range service.Build.Args {
		solveOpt.FrontendAttrs["build-arg:"+k] = v
	}

	// Build and stream progress
	pw, _ := progresswriter.NewPrinter(ctx, os.Stderr, "auto")
	res, err := socketClient.Solve(ctx, nil, solveOpt, pw.Status())

	<-pw.Done()

	if err != nil {
		return crex.UserError("build failed", "an error occurred during the build process").
			Cause(err).
			Err()
	}

	if err := pw.Err(); err != nil {
		slog.Warn("progress display error", "error", err)
	}

	// TODO: Store build metadata (digest, platforms, etc.) in accompanying file
	// Available: res.ExporterResponse["containerimage.digest"]
	//           res.ExporterResponse["containerimage.config.digest"]
	_ = res

	return nil
}

// Returns the Buildkit socket path.
//
// Rootless socket path: /run/user/<uid>/buildkit/buildkitd.sock
func buildkitSocketPath() string {
	uid := os.Getuid()
	return filepath.Join("/run/user", strconv.Itoa(uid), "buildkit", "buildkitd.sock")
}
