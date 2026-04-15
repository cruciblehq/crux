package subsystem

import (
	"path"
	"regexp"

	"github.com/cruciblehq/crex"
)

// Matches a valid capability name (uppercase letters, digits, and underscores).
var validFcapsName = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)

// Validates a Linux capability name against the kernel ABI format.
func validateFcapsName(name string) error {
	if !validFcapsName.MatchString(name) {
		return crex.Wrapf(ErrGrantExpression, "invalid capability name %q", name)
	}
	return nil
}

// Validates and normalizes an absolute container path.
func validateFcapsContainerPath(p string) (string, error) {
	p = path.Clean(p)
	if !path.IsAbs(p) {
		return "", crex.Wrapf(ErrGrantExpression, "path must be absolute %q", p)
	}
	return p, nil
}
