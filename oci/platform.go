package oci

import "strings"

// Returns the platforms required for universal deployment.
//
// Images must support all these platforms to be considered universally
// deployable within the Crucible ecosystem.
func RequiredPlatforms() []string {
	return []string{
		"linux/amd64", // x86_64 servers, most cloud providers
		"linux/arm64", // ARM servers, Apple Silicon, modern cloud instances
	}
}

// Parses a platform string into OS and architecture components.
//
// Expects the format "os/arch" (e.g., "linux/amd64", "linux/arm64"). Returns
// an error if the format is invalid. This is the inverse of the platform
// strings returned by RequiredPlatforms and Index.Platforms.
func ParsePlatform(platform string) (osName, arch string, err error) {
	parts := strings.Split(platform, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidPlatform
	}
	return parts[0], parts[1], nil
}
