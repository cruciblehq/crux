package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/pack"
	"github.com/cruciblehq/protocol/pkg/manifest"
	"github.com/cruciblehq/protocol/pkg/registry"
)

var (
	ErrPushFailed = errors.New("push failed")
)

// Pushes a resource package to the Hub registry.
type PushCmd struct {
	Hub      string `help:"Hub registry URL." default:"http://hub.cruciblehq.xyz:8080"`
	Resource string `arg:"" help:"Resource to push (namespace/name)."`
}

// Executes the push command.
func (cmd *PushCmd) Run(ctx context.Context) error {
	if err := cmd.validatePackage(); err != nil {
		return err
	}

	man, err := cmd.loadManifest()
	if err != nil {
		return err
	}

	namespace, resourceName, err := cmd.parseResource()
	if err != nil {
		return err
	}

	client := registry.NewClient(cmd.Hub, nil)

	if err := cmd.verifyNamespace(ctx, client, namespace); err != nil {
		return err
	}

	if err := cmd.ensureResource(ctx, client, namespace, resourceName, man); err != nil {
		return err
	}

	if err := cmd.createVersion(ctx, client, namespace, resourceName, man.Resource.Version); err != nil {
		return err
	}

	return cmd.uploadPackage(ctx, client, namespace, resourceName, man.Resource.Version)
}

// Validates that the package file exists.
func (cmd *PushCmd) validatePackage() error {
	if _, err := os.Stat(pack.PackageOutput); os.IsNotExist(err) {
		return crex.UserError("package not found", fmt.Sprintf("%s does not exist", pack.PackageOutput)).
			Fallback("Run 'crux pack' first to create the package.").
			Err()
	}
	return nil
}

// Loads and validates the manifest.
func (cmd *PushCmd) loadManifest() (*manifest.Manifest, error) {
	man, err := manifest.Read(pack.Manifestfile)
	if err != nil {
		return nil, crex.UserError("failed to read manifest", err.Error()).
			Fallback("Ensure crucible.yaml exists and is valid.").
			Err()
	}
	return man, nil
}

// Parses namespace and resource name from the resource argument.
func (cmd *PushCmd) parseResource() (namespace, resourceName string, err error) {
	parts := strings.Split(cmd.Resource, "/")
	if len(parts) != 2 {
		return "", "", crex.UserError("invalid resource format", "expected namespace/name").
			Fallback("Use format: crux push namespace/name").
			Err()
	}
	return parts[0], parts[1], nil
}

// Creates a new version in the registry.
func (cmd *PushCmd) createVersion(ctx context.Context, client *registry.Client, namespace, resource, version string) error {
	versionInfo := registry.VersionInfo{
		String: version,
	}

	_, err := client.CreateVersion(ctx, namespace, resource, versionInfo)
	if err != nil {
		var regErr *registry.Error
		if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeVersionExists {
			return crex.UserError("version already exists", fmt.Sprintf("version %s already exists", version)).
				Fallback("Increment the version in crucible.yaml and rebuild.").
				Err()
		}
		return crex.UserError("failed to create version", err.Error()).
			Fallback("Check Hub connectivity and permissions.").
			Err()
	}

	return nil
}

// Uploads the package archive to the registry.
func (cmd *PushCmd) uploadPackage(ctx context.Context, client *registry.Client, namespace, resource, version string) error {
	archive, err := os.Open(pack.PackageOutput)
	if err != nil {
		return crex.Wrap(ErrPushFailed, err)
	}
	defer archive.Close()

	_, err = client.UploadArchive(ctx, namespace, resource, version, archive)
	if err != nil {
		return crex.UserError("failed to upload archive", err.Error()).
			Fallback("Check Hub connectivity and package integrity.").
			Err()
	}

	return nil
}

// Verifies that the namespace exists.
func (cmd *PushCmd) verifyNamespace(ctx context.Context, client *registry.Client, namespace string) error {
	_, err := client.ReadNamespace(ctx, namespace)
	if err == nil {
		return nil // Namespace exists
	}

	var regErr *registry.Error
	if errors.As(err, &regErr) && regErr.Code == registry.ErrorCodeNotFound {
		return crex.UserError("namespace not found", fmt.Sprintf("namespace '%s' does not exist", namespace)).
			Fallback("Create the namespace first using the Hub administration tools.").
			Err()
	}

	return crex.UserError("failed to check namespace", err.Error()).
		Fallback("Check Hub connectivity.").
		Err()
}

// Ensures the resource exists, creating it if necessary.
func (cmd *PushCmd) ensureResource(ctx context.Context, client *registry.Client, namespace, resource string, man *manifest.Manifest) error {
	_, err := client.ReadResource(ctx, namespace, resource)
	if err == nil {
		return nil // Resource exists
	}

	// Check if error is "not found"
	var regErr *registry.Error
	if !errors.As(err, &regErr) || regErr.Code != registry.ErrorCodeNotFound {
		return crex.UserError("failed to check resource", err.Error()).
			Fallback("Check Hub connectivity.").
			Err()
	}

	// Create resource
	resInfo := registry.ResourceInfo{
		Name:        resource,
		Type:        man.Resource.Type,
		Description: "",
	}
	_, err = client.CreateResource(ctx, namespace, resInfo)
	if err != nil {
		return crex.UserError("failed to create resource", err.Error()).
			Fallback("Check Hub connectivity and permissions.").
			Err()
	}

	return nil
}
