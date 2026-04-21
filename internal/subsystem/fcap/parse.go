package fcap

import (
	pathpkg "path"
	"strings"

	"github.com/cruciblehq/crex"

	"github.com/cruciblehq/crux/internal/subsystem/shared"
)

// Parses an fcap rule string into a Grant.
//
// The rule has the form "path mode cap" where path is an absolute path, mode
// is "effective" or "inheritable", and cap is a capability name without the
// CAP_ prefix. Path validation requires the path to already be clean so
// repeated grants for the same executable always use one unambiguous form.
// Returns an error if the rule is malformed, or either the mode or capability
// are unknown.
func Parse(rule string) (*Grant, error) {
	fields := strings.Fields(rule)
	if len(fields) != 3 {
		return nil, crex.Wrapf(ErrInvalidRule, "expected path mode capability")
	}

	path, err := normalizePath(fields[0])
	if err != nil {
		return nil, err
	}

	mode, err := ParseMode(fields[1])
	if err != nil {
		return nil, err
	}

	name := fields[2]
	if _, err := shared.ParseCap(name); err != nil {
		return nil, crex.Wrapf(ErrInvalidRule, "unknown capability %q", name)
	}

	return &Grant{Mode: mode, Path: path, Caps: []string{name}}, nil
}

// Validates a binary path.
//
// The path must be non-empty, absolute, not have a trailing slash, contain no
// NUL bytes, and already be clean. A clean path is exactly path.Clean(path).
// Paths that require normalization are rejected to avoid ambiguous inputs.
func normalizePath(path string) (string, error) {
	if path == "" {
		return "", crex.Wrapf(ErrInvalidRule, "path is empty")
	}
	if strings.Contains(path, "\x00") {
		return "", crex.Wrapf(ErrInvalidRule, "path %q contains NUL", path)
	}
	if !pathpkg.IsAbs(path) {
		return "", crex.Wrapf(ErrInvalidRule, "path %q must be absolute", path)
	}
	if strings.HasSuffix(path, "/") {
		return "", crex.Wrapf(ErrInvalidRule, "path %q must not have a trailing slash", path)
	}

	clean := pathpkg.Clean(path)
	if path != clean {
		return "", crex.Wrapf(ErrInvalidRule, "path %q must be clean", path)
	}
	return clean, nil
}
