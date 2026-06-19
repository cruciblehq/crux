package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Represents the 'crux push' command.
type PushCmd struct {
	Registry string `help:"Hub registry URL (default: http://hub.cruciblehq.xyz:8080)."`
}

// Executes the push command.
func (c *PushCmd) Run(ctx context.Context) error {
	registryURL := c.Registry
	if registryURL == "" {
		registryURL = internal.DefaultRegistryURL
	}

	slog.Info("pushing package...", "registry", registryURL)

	src, err := hub.NewSource(registryURL, internal.DefaultNamespace)
	if err != nil {
		return err
	}

	man, err := manifest.ReadAt(RootCmd.Context)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return crex.UserError("no manifest found", "no crucible.yaml was found in the current directory").
				Recovery("Run this command from a directory containing a crucible.yaml.").
				Cause(err).
				Err()
		}
		return crex.UserError("invalid manifest", "the crucible.yaml could not be read or parsed").
			Recovery("Check that crucible.yaml is present and valid.").
			Cause(err).
			Err()
	}

	if err := resource.Push(ctx, src, *man, file.Package(RootCmd.Context)); err != nil {
		const description = "cannot push package"
		switch {
		case errors.Is(err, registry.ErrVersionExists):
			return crex.UserError(description, "the resource version already exists in the registry").
				Recovery("Bump resource.version in crucible.yaml, then run 'crux pack' and 'crux push' again.").
				Reclassify(err)
		case errors.Is(err, registry.ErrNamespaceNotFound):
			return crex.UserError(description, "the target namespace does not exist in the registry").
				Recovery("Create the namespace in the registry, then try again.").
				Reclassify(err)
		case errors.Is(err, registry.ErrInvalidIdentifier):
			return crex.UserError(description, "the resource name in crucible.yaml is invalid").
				Recovery("Use the format 'namespace/resource' in resource.name, then try again.").
				Reclassify(err)
		case errors.Is(err, registry.ErrFileSystemOperation):
			return crex.SystemError(description, "the package archive could not be read from disk").
				Recoveryf("Run 'crux pack' to create %s, then try again.", file.Package(RootCmd.Context)).
				Reclassify(err)
		case errors.Is(err, registry.ErrRegistryOperation):
			return crex.SystemError(description, "the registry could not accept the package upload").
				Recovery("Check your network connection and try again.").
				Reclassify(err)
		default:
			return err
		}
	}

	slog.Info("package pushed successfully")

	return nil
}
