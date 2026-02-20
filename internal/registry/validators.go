package registry

import (
	"regexp"

	"github.com/cruciblehq/spec/reference"
)

var (

	// Valid name pattern
	namePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
)

// Whether a name is valid for namespace, resource, channel names.
//
// Names may include lowercase letters (a–z), digits (0–9), and hyphens (-),
// must start and end with an alphanumeric character, and must not exceed 63
// characters.
func validateName(name string) error {
	if name == "" {
		return ErrNameEmpty
	}
	if len(name) > 63 {
		return ErrNameTooLong
	}
	if !namePattern.MatchString(name) {
		return ErrNameInvalid
	}
	return nil
}

// Whether a version string is valid.
//
// Uses proper semantic version validation from the reference package.
func validateVersionString(version string) error {
	if _, err := reference.ParseVersion(version); err != nil {
		return ErrVersionInvalid
	}
	return nil
}

// Validates a namespace identifier.
//
// Ensures the namespace name follows naming conventions.
func validateNamespace(namespace string) error {
	return validateName(namespace)
}

// Validates a resource identifier (namespace + resource).
//
// Ensures both namespace and resource names follow naming conventions.
func validateIdentifier(namespace, resource string) error {
	if err := validateName(namespace); err != nil {
		return err
	}
	return validateName(resource)
}

// Validates a version reference (namespace + resource + version).
//
// Ensures namespace and resource names follow naming conventions and
// the version string is a valid semantic version.
func validateReference(namespace, resource, version string) error {
	if err := validateName(namespace); err != nil {
		return err
	}
	if err := validateName(resource); err != nil {
		return err
	}
	return validateVersionString(version)
}

// Validates a channel reference (namespace + resource + channel).
//
// Ensures namespace, resource, and channel names all follow naming conventions.
func validateChannelReference(namespace, resource, channel string) error {
	if err := validateName(namespace); err != nil {
		return err
	}
	if err := validateName(resource); err != nil {
		return err
	}
	return validateName(channel)
}

// Validates channel info (namespace + resource + channel name + version).
//
// Ensures namespace, resource, and channel names follow naming conventions,
// and the target version string is a valid semantic version.
func validateChannelInfo(namespace, resource string, info ChannelInfo) error {
	if err := validateName(namespace); err != nil {
		return err
	}
	if err := validateName(resource); err != nil {
		return err
	}
	if err := validateName(info.Name); err != nil {
		return err
	}
	return validateVersionString(info.Version)
}
