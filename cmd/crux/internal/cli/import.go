package cli

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/cruciblehq/utils-go/crex"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

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
		platform = fmt.Sprintf("linux/%s", runtime.GOARCH)
	}

	p, err := v1.ParsePlatform(platform)
	if err != nil {
		return crex.UserError("invalid platform", platform).
			Recovery("Use a valid platform such as 'linux/amd64' or 'linux/arm64'.").
			Cause(err).
			Err()
	}

	parsed, err := name.ParseReference(c.Image)
	if err != nil {
		return crex.UserError("invalid image reference", c.Image).
			Recovery("Use a valid OCI image reference such as 'alpine:3.21' or 'ghcr.io/user/image:tag'.").
			Cause(err).
			Err()
	}

	slog.Info("pulling image...", "image", parsed.String(), "platform", platform)

	img, err := remote.Image(parsed, remote.WithContext(ctx), remote.WithPlatform(*p))
	if err != nil {
		return crex.Wrap(ErrImport, err)
	}

	if err := crane.Save(img, parsed.String(), c.Output); err != nil {
		return crex.Wrap(ErrImport, err)
	}

	slog.Info("image imported", "image", parsed.String(), "output", c.Output)
	return nil
}
