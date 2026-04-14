package subsystem

import (
	"path"
	"regexp"

	"github.com/cruciblehq/crex"
)

// Matches a valid capability name (uppercase letters, digits, and underscores).
var validCapName = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)

// Validates a Linux capability name against the kernel ABI format.
func validateCapName(name string) error {
	if !validCapName.MatchString(name) {
		return crex.Wrapf(ErrGrantExpression, "invalid capability name %q", name)
	}
	return nil
}

// Validates and normalizes an absolute container path.
func validateContainerPath(p string) (string, error) {
	p = path.Clean(p)
	if !path.IsAbs(p) {
		return "", crex.Wrapf(ErrGrantExpression, "path must be absolute %q", p)
	}
	return p, nil
}

// Matches a valid device name (lowercase letters, digits, and underscores).
var validDeviceName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Validates a device name against the expected format.
func validateDeviceName(name string) error {
	if !validDeviceName.MatchString(name) {
		return crex.Wrapf(ErrGrantExpression, "invalid device name %q", name)
	}
	return nil
}

// Validates that a file mode fits within the 12-bit Unix permission range (07777).
func validateFileMode(mode uint16) error {
	if mode > 0o7777 {
		return crex.Wrapf(ErrGrantExpression, "%d is not a valid file mode", mode)
	}
	return nil
}
