package resource

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/cache"
	"github.com/cruciblehq/crux/internal/registry"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
	specregistry "github.com/cruciblehq/spec/registry"
)

// Pushes a resource package to the Hub registry.
//
// The manifest is used to determine the target registry, namespace, and
// resource name. The resource name is re-resolved here using the provided
// reference options rather than read from the archive, because extracting
// a compressed archive solely to read the manifest would be wasteful. The
// resolved manifest inside the archive (written by [Builder.Build]) is
// functionally equivalent since the same options produce the same result.
func push(ctx context.Context, m manifest.Manifest, packagePath string, source Source) error {
	if _, err := os.Stat(packagePath); err != nil {
		return crex.UserError("package not found", "package does not exist").
			Fallback("Run 'crux pack' first to create the package.").
			Cause(err).
			Err()
	}

	id, err := reference.ParseIdentifier(m.Resource.Name, string(m.Resource.Type))
	if err != nil {
		return crex.UserError("invalid resource name", "could not parse the resource identifier").
			Fallback("Check the resource name in crucible.yaml.").
			Cause(err).
			Err()
	}

	client := registry.NewClient(source.Registry, nil)

	if err := verifyNamespace(ctx, client, id.Namespace()); err != nil {
		return err
	}

	if err := ensureResource(ctx, client, id.Namespace(), id.Name(), &m); err != nil {
		return err
	}

	if err := createVersion(ctx, client, id.Namespace(), id.Name(), m.Resource.Version); err != nil {
		return err
	}

	return uploadPackage(ctx, client, id.Namespace(), id.Name(), m.Resource.Version, packagePath)
}

// Verifies that the namespace exists.
//
// Makes a call to the registry to check if the specified namespace exists. If
// the namespace does not exist, an error is returned prompting the user to
// create it first.
func verifyNamespace(ctx context.Context, client *registry.Client, namespace string) error {
	_, err := client.ReadNamespace(ctx, namespace)
	if err == nil {
		return nil // Namespace exists
	}

	var regErr *specregistry.Error
	if errors.As(err, &regErr) && regErr.Code == specregistry.ErrorCodeNotFound {
		return crex.UserError("namespace not found", fmt.Sprintf("namespace '%s' does not exist", namespace)).
			Fallback("Create the namespace first.").
			Err()
	}

	return crex.UserError("failed to check namespace", "could not verify the namespace exists").
		Fallback("Crucible's Hub connectivity may be impaired. Try again later.").
		Cause(err).
		Err()
}

// Ensures the resource exists, creating it if necessary.
//
// Checks if the resource exists in the registry. If not, it creates the
// resource using the information from the manifest.
func ensureResource(ctx context.Context, client *registry.Client, namespace, resource string, man *manifest.Manifest) error {
	_, err := client.ReadResource(ctx, namespace, resource)
	if err == nil {
		return nil // Resource exists
	}

	var regErr *specregistry.Error
	if !errors.As(err, &regErr) || regErr.Code != specregistry.ErrorCodeNotFound {
		return crex.UserError("failed to check resource", "could not verify the resource exists").
			Fallback("Crucible's Hub connectivity may be impaired. Try again later.").
			Cause(err).
			Err()
	}

	resInfo := specregistry.ResourceInfo{
		Name:        resource,
		Type:        string(man.Resource.Type),
		Description: "",
	}
	_, err = client.CreateResource(ctx, namespace, resInfo)
	if err != nil {
		return crex.UserError("failed to create resource", "the resource could not be registered").
			Fallback("Check Hub connectivity and permissions.").
			Cause(err).
			Err()
	}

	return nil
}

// Creates a new version in the registry.
//
// Attempts to create the specified version for the resource. If the version
// already exists, an error is returned prompting the user to increment the
// version in the manifest.
func createVersion(ctx context.Context, client *registry.Client, namespace, resource, version string) error {
	versionInfo := specregistry.VersionInfo{
		String: version,
	}

	_, err := client.CreateVersion(ctx, namespace, resource, versionInfo)
	if err != nil {
		var regErr *specregistry.Error
		if errors.As(err, &regErr) && regErr.Code == specregistry.ErrorCodeVersionExists {
			return crex.UserError("version already exists", fmt.Sprintf("version %s already exists", version)).
				Fallback("Increment the version in crucible.yaml and rebuild.").
				Err()
		}
		return crex.UserError("failed to create version", "the version could not be registered").
			Fallback("Check Hub connectivity and permissions.").
			Cause(err).
			Err()
	}

	return nil
}

// Uploads the package archive to the registry.
//
// Opens the package file and uploads it to the specified resource version in
// the registry. After successful upload, updates the local cache.
func uploadPackage(ctx context.Context, client *registry.Client, namespace, resource, version, packageOutput string) error {
	archive, err := os.Open(packageOutput)
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	defer archive.Close()

	_, err = client.UploadArchive(ctx, namespace, resource, version, archive)
	if err != nil {
		return crex.UserError("failed to upload archive", "the package could not be uploaded to the registry").
			Fallback("Check Hub connectivity and package integrity.").
			Cause(err).
			Err()
	}

	// Update local cache with pushed package
	if err := updateLocalCache(namespace, resource, version, packageOutput); err != nil {
		// Log, but don't fail the push; the remote was updated successfully
		slog.Error("failed to update local cache", "error", err)
	}

	return nil
}

// Adds the pushed package to the local cache.
//
// This ensures the local cache is in sync with the remote after a push,
// avoiding the need to re-download the package if it's needed locally.
func updateLocalCache(namespace, resource, version, packagePath string) error {
	localCache, err := cache.Open()
	if err != nil {
		return crex.Wrap(ErrCacheOperation, err)
	}
	defer localCache.Close()

	archive, err := os.Open(packagePath)
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	defer archive.Close()

	_, err = localCache.Put(namespace, resource, version, archive)
	if err != nil {
		return crex.Wrap(ErrCacheOperation, err)
	}
	return nil
}
