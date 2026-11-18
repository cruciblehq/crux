package compute

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/spec/reference"
	"github.com/cruciblehq/utils-go/crex"
)

// Default machine image info.
const (
	machineType      = "machine"         // Registry resource type for the machine image.
	machineNamespace = "crucible"        // Registry namespace for the machine image.
	machineName      = "machine-default" // Registry resource name for the machine image.
	machineVersion   = "0.3.0"           // Pinned machine image version.
	machineExtension = ".qcow2"          // Disk image file extension.
)

// Value of [runtime.GOARCH] on 64-bit ARM.
const goarchARM64 = "arm64"

// Architecture names of the disk images in the machine image archive.
const (
	imageArchARM64 = "aarch64" // 64-bit ARM.
	imageArchAMD64 = "x86_64"  // 64-bit x86.
)

// Ensures the default machine disk image is available in the local cache.
//
// Downloads from a given registry if not already cached. Returns the local
// filesystem path to the image file.
func EnsureMachineImage(ctx context.Context, registryURL *url.URL) (string, error) {
	path, err := cachedMachineImagePath()
	if errors.Is(err, ErrMachineImageMissing) {
		slog.Info("machine image not in cache, downloading...",
			"namespace", machineNamespace,
			"name", machineName,
			"version", machineVersion,
		)
		if err := fetchMachineImage(ctx, registryURL); err != nil {
			return "", err
		}
		path, err = cachedMachineImagePath()
	}
	if err != nil {
		return "", err
	}
	return path, nil
}

// Downloads and caches the machine disk image from the registry.
//
// Pulls the pinned machine version and extracts it into the local registry
// cache. After this returns, [cachedMachineImagePath] should resolve. The
// image is built and published by the Crucible team. It's an Alpine image
// with containerd installed and used as the base for provisioning the VM.
func fetchMachineImage(ctx context.Context, registryURL *url.URL) error {
	const description = "cannot download machine image"
	if registryURL == nil {
		return crex.ProgrammingError(description, "the machine image registry URL is required").
			Err()
	}

	registry := registryURL.String()
	src, err := hub.NewSource(registry, machineNamespace)
	if err != nil {
		return crex.SystemError(description, "the registry source could not be initialized").
			Recovery("Check your network connection and try again.").
			Cause(err).
			Err()
	}

	id := reference.NewIdentifier(machineType, registry, machineNamespace, machineName)
	ref, err := reference.New(id, machineVersion, nil)
	if err != nil {
		return crex.SystemError(description, "the machine image reference is invalid").
			Recovery("If the problem persists, report it to the Crucible team.").
			Cause(err).
			Err()
	}

	if _, err := src.Pull(ctx, ref); err != nil {
		return crex.SystemError(description, "the machine image could not be retrieved from the registry").
			Recovery("Check your network connection and try again.").
			Cause(err).
			Err()
	}
	return nil
}

// Returns the local path of the cached machine disk image.
//
// Checks the expected cache location for the image file. If the file is not
// found, returns [ErrMachineImageMissing]. It's up to the caller to decide
// whether to attempt a fetch from the registry.
func cachedMachineImagePath() (string, error) {
	c, err := cache.Open()
	if err != nil {
		return "", err
	}
	defer c.Close()

	ok, err := c.HasExtracted(machineNamespace, machineName, machineVersion)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", crex.Newf(ErrMachineImageMissing, "%s/%s %s is not in the cache", machineNamespace, machineName, machineVersion)
	}
	dir, err := c.Extract(machineNamespace, machineName, machineVersion)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s%s", machineArch(), machineExtension))
	if _, err := os.Stat(path); err != nil {
		return "", crex.Newf(ErrMachineImageMissing, "expected at %s", path)
	}
	return path, nil
}

// Returns the architecture name the machine image archive uses for this
// platform (x86_64 or aarch64).
func machineArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return imageArchARM64
	default:
		return imageArchAMD64
	}
}
