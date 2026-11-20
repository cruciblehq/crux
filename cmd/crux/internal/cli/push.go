package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/utils-go/crex"
)

// Description of errors from pushing a package.
const descPushPackage = "cannot push package"

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

	man, err := readProjectManifest()
	if err != nil {
		return err
	}

	pkg := resource.PackagePath(RootCmd.Context)
	archive, err := os.Open(pkg)
	if err != nil {
		return crex.SystemError(descPushPackage, "the package archive could not be read from disk").
			Recoveryf("Run 'crux pack' to create %s, then try again.", pkg).
			Cause(err).
			Err()
	}
	defer archive.Close()

	if err := resource.Push(ctx, src, *man, archive); err != nil {
		return pushError(err)
	}

	slog.Info("package pushed successfully")

	return nil
}

// Maps a push failure to a user error.
func pushError(err error) error {
	switch {
	case errors.Is(err, hub.ErrVersionExists):
		return crex.UserError(descPushPackage, "the resource version already exists in the registry").
			Recovery("Bump resource.version in crucible.yaml, then run 'crux pack' and 'crux push' again.").
			Reclassify(err)
	case errors.Is(err, hub.ErrNamespaceNotFound):
		return crex.UserError(descPushPackage, "the target namespace does not exist in the registry").
			Recovery("Create the namespace in the registry, then try again.").
			Reclassify(err)
	case errors.Is(err, hub.ErrInvalidIdentifier):
		return crex.UserError(descPushPackage, "the resource name in crucible.yaml is invalid").
			Recovery("Use the format 'namespace/resource' in resource.name, then try again.").
			Reclassify(err)
	case errors.Is(err, hub.ErrRegistryOperation):
		return crex.SystemError(descPushPackage, "the registry could not accept the package upload").
			Recovery("Check your network connection and try again.").
			Reclassify(err)
	default:
		return err
	}
}
