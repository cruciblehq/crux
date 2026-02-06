package reference

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/kit/crex"
)

// Represents a parsed semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

var (
	// Prerelease: identifier.number (e.g., "alpha.1", "beta.2", "rc.3").
	// Identifier must start with a letter and contain only alphanumerics.
	// Number must be zero or a positive integer without leading zeros.
	prereleasePattern = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9]*)\.([0-9]|[1-9][0-9]+)$`)

	// Build: alphanumeric and hyphens, dot-separated.
	buildPattern = regexp.MustCompile(`^[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*$`)
)

// Parses a semantic version string.
//
// Versions must include major, minor, and patch components with optional "v"
// or "V" prefix. Prereleases must follow the format "identifier.number"
// (e.g., "alpha.1", "beta.2", "rc.3").
func ParseVersion(version string) (*Version, error) {
	s := strings.TrimSpace(version)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")

	v := &Version{}

	// Extract build metadata
	if idx := strings.Index(s, "+"); idx != -1 {
		v.Build = s[idx+1:]
		s = s[:idx]

		if v.Build == "" || !buildPattern.MatchString(v.Build) {
			return nil, crex.Wrap(ErrInvalidReference, ErrInvalidBuildMetadata)
		}
	}

	// Extract and validate prerelease
	if idx := strings.Index(s, "-"); idx != -1 {
		v.Prerelease = s[idx+1:]
		s = s[:idx]

		if !prereleasePattern.MatchString(v.Prerelease) {
			return nil, crex.Wrap(ErrInvalidReference, ErrInvalidPrereleaseFormat)
		}
	}

	// Parse version numbers
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, crex.Wrap(ErrInvalidReference, ErrInvalidVersionComponents)
	}

	var err error
	if v.Major, err = strconv.Atoi(parts[0]); err != nil || v.Major < 0 {
		return nil, crex.Wrap(ErrInvalidReference, ErrInvalidMajorVersion)
	}

	if v.Minor, err = strconv.Atoi(parts[1]); err != nil || v.Minor < 0 {
		return nil, crex.Wrap(ErrInvalidReference, ErrInvalidMinorVersion)
	}

	if v.Patch, err = strconv.Atoi(parts[2]); err != nil || v.Patch < 0 {
		return nil, crex.Wrap(ErrInvalidReference, ErrInvalidPatchVersion)
	}

	return v, nil
}

// Returns the canonical string representation (e.g., "1.2.3-alpha.1+001").
func (v *Version) String() string {
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(v.Major))
	sb.WriteString(".")
	sb.WriteString(strconv.Itoa(v.Minor))
	sb.WriteString(".")
	sb.WriteString(strconv.Itoa(v.Patch))
	if v.Prerelease != "" {
		sb.WriteString("-")
		sb.WriteString(v.Prerelease)
	}
	if v.Build != "" {
		sb.WriteString("+")
		sb.WriteString(v.Build)
	}
	return sb.String()
}

// Returns true if this is a prerelease version.
func (v *Version) IsPrerelease() bool {
	return v.Prerelease != ""
}

// Compares two versions.
//
// Returns -1 if v < other, 0 if v == other, 1 if v > other. The second return
// value indicates whether the comparison is valid.
//
// Comparison is invalid when both versions have different prerelease
// identifiers (e.g., "alpha" vs "beta"). Versions with the same identifier
// are compared by their prerelease number (e.g., "alpha.1" < "alpha.2").
//
// A stable version is always greater than a prerelease of the same
// major.minor.patch (e.g., "1.0.0" > "1.0.0-alpha.1").
//
// Build metadata is ignored for comparison purposes.
func (v *Version) Compare(other *Version) (int, bool) {
	if c := compareInt(v.Major, other.Major); c != 0 {
		return c, true
	}
	if c := compareInt(v.Minor, other.Minor); c != 0 {
		return c, true
	}
	if c := compareInt(v.Patch, other.Patch); c != 0 {
		return c, true
	}

	return comparePrerelease(v.Prerelease, other.Prerelease)
}

// Compares two integers.
func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// Compares prerelease strings.
//
// Returns (comparison, valid). Valid is false when prereleases have different
// identifiers (e.g., "alpha" vs "beta").
//
// Stable (no prerelease) is always greater than prerelease.
func comparePrerelease(a, b string) (int, bool) {
	if a == "" && b == "" {
		return 0, true
	}
	if a == "" {
		return 1, true // stable > prerelease
	}
	if b == "" {
		return -1, true // prerelease < stable
	}

	aID, aNum := splitPrerelease(a)
	bID, bNum := splitPrerelease(b)

	if aID != bID {
		return 0, false
	}

	return compareInt(aNum, bNum), true
}

// Splits a validated prerelease into identifier and number.
//
// Assumes the prerelease has been validated by prereleasePattern.
//
// Note: these fields are not stored separately in Version to avoid direct
// comparisons, which could be tempting but incorrect.
func splitPrerelease(prerelease string) (string, int) {
	idx := strings.LastIndex(prerelease, ".")
	id := prerelease[:idx]
	num, _ := strconv.Atoi(prerelease[idx+1:])
	return id, num
}
