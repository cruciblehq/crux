package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	goruntime "runtime"

	"github.com/cruciblehq/crex"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

var errImport = errors.New("import failed")

// Represents the 'crux import' command.
type ImportCmd struct {
	Image    string `arg:"" help:"OCI image reference (e.g., alpine:3.21)."`
	Output   string `short:"o" required:"" help:"Output file path for the OCI archive."`
	Platform string `short:"p" help:"Target platform (e.g., linux/amd64). Defaults to the host platform."`
}

// Pulls a remote OCI image and saves it as a local OCI archive.
func (c *ImportCmd) Run(ctx context.Context) error {
	platform := c.Platform
	if platform == "" {
		platform = fmt.Sprintf("linux/%s", goruntime.GOARCH)
	}

	p, err := v1.ParsePlatform(platform)
	if err != nil {
		return crex.Wrap(errImport, err)
	}

	parsed, err := name.ParseReference(c.Image)
	if err != nil {
		return crex.Wrap(errImport, err)
	}

	slog.Info("pulling image...", "image", parsed.String(), "platform", platform)

	img, err := remote.Image(parsed, remote.WithPlatform(*p))
	if err != nil {
		return crex.Wrap(errImport, err)
	}

	f, err := os.Create(c.Output)
	if err != nil {
		return crex.Wrap(errImport, err)
	}
	defer f.Close()

	if err := crane.Save(img, parsed.String(), c.Output); err != nil {
		return crex.Wrap(errImport, err)
	}

	slog.Info("image imported", "image", parsed.String(), "output", c.Output)
	return nil
}
