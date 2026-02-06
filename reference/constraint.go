package reference

import (
	"strconv"
	"strings"
)

// Represents a single version comparison rule.
//
// A constraint consists of an operator and a version specification. The version
// may be partial (e.g., "1.2", without patch) in which case minorSet/patchSet
// track which components were explicitly specified.
//
// The operator defines how to compare versions against this constraint.
// Supported operators are:
//   - "=": exact match (partial constraints match any sub-version)
//   - "!=": negated exact match
//   - ">": greater than
//   - "<": less than
//   - ">=": greater-or-equal
//   - "<=": less-or-equal
//   - "~": tilde range, allows patch-level changes
//   - "^": caret range, allows minor-level changes within major
//
// Pre-release and build metadata are not supported in constraints.
type constraint struct {
	operator string // The comparison operator (e.g., "=", ">=", "~", etc.)
	major    int    // The major version (always set).
	minor    int    // The minor version (only valid if minorSet is true).
	patch    int    // The patch version (only valid if patchSet is true).
	minorSet bool   // Whether the minor version is set.
	patchSet bool   // Whether the patch version is set.
}

// Whether a version satisfies this constraint.
//
// Prerelease versions never match constraints. A prerelease like "1.2.3-alpha"
// must be explicitly selected via a channel, not a version constraint.
func (c constraint) matches(v *Version) bool {
	if v.Prerelease != "" {
		return false
	}

	switch c.operator {
	case "=":
		return c.matchEqual(v)
	case "!=":
		return !c.matchEqual(v)
	case ">":
		return c.compare(v) < 0
	case ">=":
		return c.compare(v) <= 0
	case "<":
		return c.compare(v) > 0
	case "<=":
		return c.compare(v) >= 0
	case "~":
		return c.matchTilde(v)
	case "^":
		return c.matchCaret(v)
	}
	return false
}

// Whether a version satisfies a tilde constraint.
//
// Tilde constraints allow patch-level changes while pinning major and minor
// versions. The behavior depends on which components are set:
//   - ~1 matches any 1.x.x (only major specified)
//   - ~1.2 matches any 1.2.x (major and minor specified)
//   - ~1.2.3 matches 1.2.x where x >= 3 (all components specified)
func (c constraint) matchTilde(v *Version) bool {
	if v.Major != c.major {
		return false
	}
	if !c.minorSet {
		return true // ~1 matches any 1.x.x
	}
	if v.Minor != c.minor {
		return false
	}
	if !c.patchSet {
		return true // ~1.2 matches any 1.2.x
	}
	return v.Patch >= c.patch
}

// Whether a version satisfies a caret constraint.
//
// Caret constraints allow changes that don't modify the leftmost non-zero
// component. The behavior differs based on whether major is zero. When major
// is non-zero, both minor and patch changes are allowed; when major is zero,
// only patch changes are allowed within the specified minor version.
func (c constraint) matchCaret(v *Version) bool {
	if v.Major != c.major {
		return false
	}
	if c.major == 0 {
		return c.matchCaretMajorZero(v)
	}
	return c.matchCaretMajorNonZero(v)
}

// Handles caret matching when major version is 0.
//
// Per semver, when major is 0, the caret operator is more restrictive because
// 0.x.y versions are considered unstable:
//   - ^0 matches any 0.x.x
//   - ^0.0 matches any 0.0.x
//   - ^0.0.3 matches exactly 0.0.3 (no changes allowed)
//   - ^0.2 matches any 0.2.x
//   - ^0.2.3 matches 0.2.x where x >= 3 (patch changes only)
func (c constraint) matchCaretMajorZero(v *Version) bool {
	if !c.minorSet {
		return true // ^0 matches any 0.x.x
	}
	if v.Minor != c.minor {
		return false
	}
	if c.minor == 0 {
		// ^0.0 or ^0.0.x: patch must match exactly if set
		if !c.patchSet {
			return true // ^0.0 matches any 0.0.x
		}
		return v.Patch == c.patch // ^0.0.3 matches only 0.0.3
	}
	// ^0.y where y > 0: allow patch-level changes
	if !c.patchSet {
		return true // ^0.2 matches any 0.2.x
	}
	return v.Patch >= c.patch // ^0.2.3 matches 0.2.x where x >= 3
}

// Handles caret matching when major version is non-zero.
//
// Per semver, when major is non-zero, the caret allows minor and patch changes:
//   - ^1 matches any 1.x.x
//   - ^1.2 matches any 1.x.y where x >= 2
//   - ^1.2.3 matches 1.x.y where (x > 2) or (x == 2 and y >= 3)
func (c constraint) matchCaretMajorNonZero(v *Version) bool {
	if !c.minorSet {
		return true // ^1 matches any 1.x.x
	}
	if v.Minor > c.minor {
		return true // ^1.2 allows 1.3.x, 1.4.x, etc.
	}
	if v.Minor < c.minor {
		return false
	}
	// v.Minor == c.minor
	if !c.patchSet {
		return true // ^1.2 matches any 1.2.x
	}
	return v.Patch >= c.patch // ^1.2.3 matches 1.2.x where x >= 3
}

// Checks for equality.
//
// A version matches the constraint if all set components are equal, accounting
// for unset minor/patch versions in the constraint. For example, "1.2" matches
// "1.2.x" and "1.2.3", but not "1.3.0".
func (c constraint) matchEqual(v *Version) bool {
	return c.compare(v) == 0
}

// Returns the relative ordering between the constraint and a version.
//
// Returns:
//   - -1 if the constraint version is less than v
//   - 0 if the constraint version equals v (for set components)
//   - 1 if the constraint version is greater than v
//
// The comparison algorithm only considers components that are explicitly set
// in the constraint. If minorSet is false, minor versions are not compared.
// Similarly, if patchSet is false, patch versions are not compared. This
// matches partial constraints like "1.2" to explicit versions like "1.2.x".
func (c constraint) compare(v *Version) int {

	if c.major != v.Major {
		if c.major < v.Major {
			return -1
		}
		return 1
	}

	if c.minorSet && c.minor != v.Minor {
		if c.minor < v.Minor {
			return -1
		}
		return 1
	}

	if c.patchSet && c.patch != v.Patch {
		if c.patch < v.Patch {
			return -1
		}
		return 1
	}

	return 0
}

// Returns the canonical string representation of the constraint.
func (c constraint) String() string {
	var b strings.Builder

	b.WriteString(c.operator)
	b.WriteString(strconv.Itoa(c.major))

	if c.minorSet {
		b.WriteByte('.')
		b.WriteString(strconv.Itoa(c.minor))
	}

	if c.patchSet {
		b.WriteByte('.')
		b.WriteString(strconv.Itoa(c.patch))
	}

	return b.String()
}
