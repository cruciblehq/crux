package manifest

import "regexp"

// Matches a valid identifier used for names (param, stage, environment).
//
// Accepts lowercase letters, digits, and underscores, starting with a letter.
// Examples: "foo", "bar_baz", "x1".
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Matches a valid platform selector in os/arch format.
//
// Accepts the standard os/arch[/variant] format used by OCI and Docker.
// Examples: "linux/amd64", "linux/arm64", "linux/arm64/v8".
var platformPattern = regexp.MustCompile(`^[a-z][a-z0-9]*/[a-z][a-z0-9_/]*$`)

// Matches a valid URL path for a route pattern.
//
// Must start with a forward slash and contain only unreserved path characters:
// letters, digits, hyphens, underscores, dots, tildes, and forward slashes.
// Query strings, fragments, and consecutive slashes are not permitted.
var routePatternRe = regexp.MustCompile(`^/[a-zA-Z0-9\-._~/]*$`)

// Whether s is a valid name identifier.
func isValidName(s string) bool {
	return namePattern.MatchString(s)
}

// Whether s is a valid platform selector.
func isValidPlatform(s string) bool {
	return platformPattern.MatchString(s)
}

// Whether s is a valid route pattern.
//
// Returns false for empty strings, patterns containing "//" or do not start
// with "/", and patterns containing query or fragment markers.
func isValidRoutePattern(s string) bool {
	if !routePatternRe.MatchString(s) {
		return false
	}
	// Disallow consecutive slashes (e.g. "//foo").
	for i := 1; i < len(s); i++ {
		if s[i] == '/' && s[i-1] == '/' {
			return false
		}
	}
	return true
}
