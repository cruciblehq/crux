package registry

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/reference"
)

// Pushes a resource package to the Hub registry.
//
// Verifies the package file exists, resolves the resource identifier, ensures
// the namespace, resource, and version exist in the registry, then uploads
// the archive.
func push(ctx context.Context, registryURL, name, resourceType, version, packagePath string) error {
	if _, err := os.Stat(packagePath); err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}

	id, err := reference.ParseIdentifier(name, resourceType)
	if err != nil {
		return crex.Wrap(ErrInvalidIdentifier, err)
	}

	client := NewClient(registryURL, nil)

	if err := verifyNamespace(ctx, client, id.Namespace()); err != nil {
		return err
	}

	if err := ensureResource(ctx, client, id.Namespace(), id.Name(), resourceType); err != nil {
		return err
	}

	if err := createVersion(ctx, client, id.Namespace(), id.Name(), version); err != nil {
		return err
	}

	return uploadPackage(ctx, client, id.Namespace(), id.Name(), version, packagePath)
}

// Checks that the namespace exists in the registry.
//
// Returns [ErrNamespaceNotFound] on a 404 response, or [ErrRegistryOperation]
// for other failures.
func verifyNamespace(ctx context.Context, client *Client, namespace string) error {
	_, err := client.ReadNamespace(ctx, namespace)
	if err == nil {
		return nil
	}

	var regErr *Error
	if errors.As(err, &regErr) && regErr.Code == ErrorCodeNotFound {
		return crex.Wrap(ErrNamespaceNotFound, err)
	}

	return crex.Wrap(ErrRegistryOperation, err)
}

// Ensures the resource exists, creating it if necessary.
//
// Looks up the resource and creates it with the given type if it does not
// yet exist.
func ensureResource(ctx context.Context, client *Client, namespace, resource, resourceType string) error {
	_, err := client.ReadResource(ctx, namespace, resource)
	if err == nil {
		return nil
	}

	var regErr *Error
	if !errors.As(err, &regErr) || regErr.Code != ErrorCodeNotFound {
		return crex.Wrap(ErrRegistryOperation, err)
	}

	resInfo := ResourceInfo{
		Name: resource,
		Type: resourceType,
	}
	_, err = client.CreateResource(ctx, namespace, resInfo)
	if err != nil {
		return crex.Wrap(ErrRegistryOperation, err)
	}

	return nil
}

// Creates a new version in the registry.
//
// Attempts to create the specified version for the resource. Returns
// [ErrVersionExists] if the version already exists.
func createVersion(ctx context.Context, client *Client, namespace, resource, version string) error {
	versionInfo := VersionInfo{
		String: version,
	}

	_, err := client.CreateVersion(ctx, namespace, resource, versionInfo)
	if err != nil {
		var regErr *Error
		if errors.As(err, &regErr) && regErr.Code == ErrorCodeVersionExists {
			return crex.Wrap(ErrVersionExists, err)
		}
		return crex.Wrap(ErrRegistryOperation, err)
	}

	return nil
}

// Uploads the package archive to the registry.
//
// Opens the package file and uploads it to the specified resource version in
// the registry. After a successful upload, updates the local cache.
func uploadPackage(ctx context.Context, client *Client, namespace, resource, version, packageOutput string) error {
	archive, err := os.Open(packageOutput)
	if err != nil {
		return crex.Wrap(ErrFileSystemOperation, err)
	}
	defer archive.Close()

	uploaded, err := client.UploadArchive(ctx, namespace, resource, version, archive)
	if err != nil {
		return crex.Wrap(ErrRegistryOperation, err)
	}

	// Cache errors are just logged, since the upload succeeded.
	if err := updateLocalCache(namespace, resource, version, packageOutput, uploaded.Digest); err != nil {
		slog.Error("failed to update local cache", "error", err)
	}

	return nil
}

// Adds the pushed package to the local cache.
//
// Stores the archive under the given namespace/resource/version with digest
// verification, keeping the local cache in sync with the remote so a
// subsequent pull does not re-download it.
func updateLocalCache(namespace, resource, version, packagePath string, digest *string) error {
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

	if digest != nil && *digest != "" {
		_, err = localCache.PutWithDigest(namespace, resource, version, *digest, archive)
	} else {
		_, err = localCache.Put(namespace, resource, version, archive)
	}
	if err != nil {
		return crex.Wrap(ErrCacheOperation, err)
	}
	return nil
}
