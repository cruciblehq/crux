package files

import (
	"path"
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Validates a slash-separated absolute path.
//
// The path must be non-empty, contain no NUL bytes, have no trailing slash, be
// absolute, and already be cleaned. Returns the cleaned path, which equals the
// input on success. Validation uses POSIX slash semantics and is independent of
// the host operating system. Wraps ErrInvalidPath on failure.
func ValidateAbsPath(p string) (string, error) {
	if p == "" {
		return "", crex.Newf(ErrInvalidPath, "path is empty")
	}
	if strings.Contains(p, "\x00") {
		return "", crex.Newf(ErrInvalidPath, "path %q contains NUL", p)
	}
	if !path.IsAbs(p) {
		return "", crex.Newf(ErrInvalidPath, "path %q must be absolute", p)
	}
	if strings.HasSuffix(p, "/") {
		return "", crex.Newf(ErrInvalidPath, "path %q must not have a trailing slash", p)
	}
	clean := path.Clean(p)
	if p != clean {
		return "", crex.Newf(ErrInvalidPath, "path %q must be clean", p)
	}
	return clean, nil
}
