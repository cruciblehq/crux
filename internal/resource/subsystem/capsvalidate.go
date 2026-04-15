package subsystem

import (
	"regexp"

	"github.com/cruciblehq/crex"
)

// Matches a valid capability name (uppercase letters, digits, and underscores).
var validCapsName = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)

// Validates a Linux capability name against the kernel ABI format.
func validateCapsName(name string) error {
	if !validCapsName.MatchString(name) {
		return crex.Wrapf(ErrGrantExpression, "invalid capability name %q", name)
	}
	return nil
}
