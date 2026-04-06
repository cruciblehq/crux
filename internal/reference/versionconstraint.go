package reference

import (
	"regexp"
	"strconv"
	"strings"
)

var (

	// Matches version numbers with optional prefix and prerelease. Prerelease
	// version identifiers are allowed syntactically, but maybe not semantically.
	versionPattern = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([a-zA-Z0-9.-]+))?$`)

	// Matches version operators.
	operatorPattern = regexp.MustCompile(`^(>=|<=|!=|>|<|=|~|\^)?(.+)$`)
)

// Represents a semantic version constraint.
//
// A version constraint defines a set of acceptable versions using comparison
// operators, ranges, and wildcards. Multiple constraint groups can be combined
// with OR logic (|| operator), while constraints within a group use AND logic
// (space-separated).
type VersionConstraint struct {
	constraints []constraintGroup
}

// Returns the canonical string representation.
//
// The canonical form normalizes whitespace and expands shorthand notations.
// Note that the order of OR groups is preserved from parsing, so semantically
// equivalent constraints may produce different strings.
func (vc *VersionConstraint) String() string {
	if vc == nil || len(vc.constraints) == 0 {
		return ""
	}
	parts := make([]string, len(vc.constraints))
	for i, g := range vc.constraints {
		parts[i] = g.String()
	}
	return strings.Join(parts, " || ")
}

// Checks whether a specific version satisfies this constraint.
//
// Parses the version string and checks it against all constraint groups.
// Returns true if any group matches (OR logic between groups).
func (vc *VersionConstraint) Matches(version string) (bool, error) {
	v, err := ParseVersion(version)
	if err != nil {
		return false, err
	}

	return vc.MatchesVersion(v)
}

// Checks whether a specific Version satisfies this constraint.
//
// Checks the Version against all constraint groups. Returns true if any group
// matches (OR logic between groups).
func (vc *VersionConstraint) MatchesVersion(v *Version) (bool, error) {
	for _, group := range vc.constraints {
		if group.matches(v) {
			return true, nil
		}
	}

	return false, nil
}

// Intersects this constraint with another, returning a new constraint that
// matches only versions satisfying both.
//
// The intersection is computed by combining each constraint group from this
// constraint with each group from the other using AND logic, then joining
// all combinations with OR logic.
//
// For example, if this = "(>=1.0.0 <2.0.0) || (>=3.0.0 <4.0.0)" and
// other = "(>=1.5.0 <3.5.0)", the result would be:
// "(>=1.5.0 <2.0.0) || (>=3.0.0 <3.5.0)"
//
// Returns an error if the intersection is empty (no versions satisfy both
// constraints). This can happen when the constraints are incompatible.
func (vc *VersionConstraint) Intersect(other *VersionConstraint) (*VersionConstraint, error) {
	if vc == nil || other == nil {
		return nil, ErrNilConstraint
	}

	var intersectedGroups []constraintGroup
	for _, g1 := range vc.constraints {
		for _, g2 := range other.constraints {
			combined := constraintGroup{
				constraints: append(append([]constraint{}, g1.constraints...), g2.constraints...),
			}

			if err := validateConstraintGroup(combined); err != nil {
				continue // Skip invalid combinations rather than failing entirely
			}

			intersectedGroups = append(intersectedGroups, combined)
		}
	}

	if len(intersectedGroups) == 0 {
		return nil, ErrIncompatibleConstraints
	}

	return &VersionConstraint{constraints: intersectedGroups}, nil
}

// Parses a version constraint string.
//
// Supports exact versions (1.2.3, =1.2.3), comparison operators (>, >=, <, <=,
// !=), hyphen ranges (1.2.3 - 2.0.0), wildcards (1.x, 1.2.x), tilde constraints
// (~1.2.3 for patch-level changes), caret constraints (^1.2.3 for minor-level
// changes), and logical operators (|| for OR, space for AND).
//
// Hyphen ranges are expanded during parsing into >= and <= constraints. For
// example, "1.2.3 - 2.0.0" becomes ">=1.2.3 <=2.0.0".
//
// Crucible requires all version ranges to have explicit upper bounds. Unbounded
// constraints like ">=1.0.0" or ">1.0.0" are rejected unless paired with an
// upper bound such as "<2.0.0".
func ParseVersionConstraint(s string) (*VersionConstraint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, wrap(ErrInvalidReference, ErrEmptyConstraint)
	}

	vc := &VersionConstraint{}

	orParts := strings.Split(s, "||")
	for _, orPart := range orParts {
		orPart = strings.TrimSpace(orPart)
		if orPart == "" {
			return nil, wrap(ErrInvalidReference, ErrEmptyOrExpression)
		}

		group, err := parseConstraintGroup(orPart)
		if err != nil {
			return nil, wrap(ErrInvalidReference, err)
		}
		vc.constraints = append(vc.constraints, group)
	}

	if len(vc.constraints) == 0 {
		return nil, wrap(ErrInvalidReference, ErrEmptyConstraint)
	}

	return vc, nil
}

// Parses space-separated AND constraints.
//
// Hyphen ranges are detected and expanded into >= and <= constraints. Multiple
// hyphen ranges and combinations with other constraints are supported. For
// example, "1.0.0 - 2.0.0 3.0.0 - 4.0.0 !=1.5.0" is valid.
func parseConstraintGroup(s string) (constraintGroup, error) {
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return constraintGroup{}, wrap(ErrInvalidReference, ErrEmptyConstraintGroup)
	}

	constraints, err := parseTokens(tokens)
	if err != nil {
		return constraintGroup{}, err
	}

	group := constraintGroup{constraints: constraints}

	if err := validateConstraintGroup(group); err != nil {
		return constraintGroup{}, err
	}

	return group, nil
}

// Whether a token is a hyphen range operator.
func isHyphenOperator(token string) bool {
	return token == "-"
}

// Parses tokens into constraints, expanding hyphen ranges.
//
// When a hyphen token is encountered, the preceding token becomes the lower
// bound (>=) and the following token becomes the upper bound (<=). Multiple
// hyphen ranges are supported.
func parseTokens(tokens []string) ([]constraint, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	if err := validateHyphenPositions(tokens); err != nil {
		return nil, err
	}

	var constraints []constraint
	i := 0

	for i < len(tokens) {
		if i+2 < len(tokens) && isHyphenOperator(tokens[i+1]) {
			lower, err := parseRangeBound(">=", tokens[i])
			if err != nil {
				return nil, wrap(ErrInvalidReference, ErrInvalidRangeBound)
			}

			upper, err := parseRangeBound("<=", tokens[i+2])
			if err != nil {
				return nil, wrap(ErrInvalidReference, ErrInvalidRangeBound)
			}

			constraints = append(constraints, lower, upper)
			i += 3
			continue
		}

		c, err := parseConstraint(tokens[i])
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, c)
		i++
	}

	return constraints, nil
}

// Validates that hyphen tokens are in valid positions.
//
// A hyphen must have a version token before and after it. Consecutive hyphens,
// leading hyphens, and trailing hyphens are invalid.
func validateHyphenPositions(tokens []string) error {
	for i, token := range tokens {
		if !isHyphenOperator(token) {
			continue
		}

		if i == 0 {
			return wrap(ErrInvalidReference, ErrLeadingHyphen)
		}

		if i == len(tokens)-1 {
			return wrap(ErrInvalidReference, ErrTrailingHyphen)
		}

		if isHyphenOperator(tokens[i-1]) || isHyphenOperator(tokens[i+1]) {
			return wrap(ErrInvalidReference, ErrConsecutiveHyphens)
		}

		if startsWithOperator(tokens[i-1]) {
			return wrap(ErrInvalidReference, ErrHyphenWithOperator)
		}

		if startsWithOperator(tokens[i+1]) {
			return wrap(ErrInvalidReference, ErrHyphenWithOperator)
		}
	}

	return nil
}

// Whether a string starts with a version operator.
func startsWithOperator(s string) bool {
	if len(s) == 0 {
		return false
	}
	switch s[0] {
	case '>', '<', '=', '!', '~', '^':
		return true
	}
	return false
}

// Parses a version as a range bound with the given operator.
//
// Range bounds must be bare versions without operators. Wildcards are not
// allowed in range bounds.
func parseRangeBound(op, version string) (constraint, error) {
	if startsWithOperator(version) {
		return constraint{}, wrap(ErrInvalidReference, ErrRangeBoundWithOperator)
	}

	if strings.Contains(version, "x") || strings.Contains(version, "X") {
		return constraint{}, wrap(ErrInvalidReference, ErrRangeBoundWithWildcard)
	}

	return parseSingleConstraint(op, version)
}

// Parses a single constraint expression.
//
// Handles bare versions (1.2.3), versions with operators (>=1.0.0), and
// wildcard versions (1.x, 1.2.x). Bare wildcards (*) are not allowed.
func parseConstraint(s string) (constraint, error) {
	if s == "*" {
		return constraint{}, wrap(ErrInvalidReference, ErrBareWildcard)
	}

	// Handle x wildcards
	if strings.HasSuffix(s, ".x") || strings.HasSuffix(s, ".X") {
		return parseWildcard(s)
	}

	// Extract operator
	match := operatorPattern.FindStringSubmatch(s)
	if match == nil {
		return constraint{}, wrap(ErrInvalidReference, ErrInvalidConstraintOperator)
	}

	op := match[1]
	version := match[2]

	if op == "" {
		op = "="
	}

	return parseSingleConstraint(op, version)
}

// Parses a version string with the given operator.
//
// If the version contains wildcards (x or X), delegates to parseWildcard.
// Otherwise, extracts major, minor, patch, and prerelease components.
func parseSingleConstraint(op, version string) (constraint, error) {
	c := constraint{operator: op}

	if strings.Contains(version, "x") || strings.Contains(version, "X") {
		return parseWildcard(op + version)
	}

	match := versionPattern.FindStringSubmatch(version)
	if match == nil {
		return c, wrap(ErrInvalidReference, ErrInvalidVersionFormat)
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return c, wrap(ErrInvalidReference, ErrInvalidMajorVersion)
	}
	c.major = major

	if match[2] != "" {
		minor, err := strconv.Atoi(match[2])
		if err != nil {
			return c, wrap(ErrInvalidReference, ErrInvalidMinorVersion)
		}
		c.minor = minor
		c.minorSet = true
	}

	if match[3] != "" {
		patch, err := strconv.Atoi(match[3])
		if err != nil {
			return c, wrap(ErrInvalidReference, ErrInvalidPatchVersion)
		}
		c.patch = patch
		c.patchSet = true
	}

	if match[4] != "" {
		return c, wrap(ErrInvalidReference, ErrPrereleaseInConstraint)
	}

	return c, nil
}

// Validates that a wildcard expression has the correct format.
//
// Valid formats are "1.x" and "1.2.x", where x represents any value for that
// component and all subsequent components. Invalid formats include bare
// wildcards (x, X), multiple wildcards (1.x.x), and wildcards not at the end
// (x.1).
func validateWildcardFormat(s string) error {

	if s == "" || s == "x" || s == "X" {
		return wrap(ErrInvalidReference, ErrBareWildcard)
	}

	xCount := strings.Count(s, ".x") + strings.Count(s, ".X")
	dotCount := strings.Count(s, ".")

	if xCount > 1 || (xCount == 1 && dotCount > 2) {
		return wrap(ErrInvalidReference, ErrMultipleWildcards)
	}

	return nil
}

// Parses a wildcard version.
//
// Wildcards are treated as equality constraints on the specified components.
// For example, "1.x" matches any version with major version 1, and "1.2.x"
// matches any version with major 1 and minor 2. Operators are not allowed
// with wildcards.
func parseWildcard(s string) (constraint, error) {
	op := "="
	if len(s) > 0 {
		switch s[0] {
		case '>', '<', '!', '~', '^':
			return constraint{}, wrap(ErrInvalidReference, ErrWildcardWithOperator)
		case '=':
			s = s[1:] // Strip explicit = prefix
		}
	}

	// Validate
	if err := validateWildcardFormat(s); err != nil {
		return constraint{}, err
	}

	// Remove trailing .x (validation ensures only one)
	s = strings.TrimSuffix(s, ".x")
	s = strings.TrimSuffix(s, ".X")

	// Parse major and optional minor
	parts := strings.Split(s, ".")
	c := constraint{operator: op}

	if len(parts) >= 1 && parts[0] != "" {
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return c, wrap(ErrInvalidReference, ErrInvalidMajorVersion)
		}
		c.major = major
	}

	if len(parts) >= 2 && parts[1] != "" {
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return c, wrap(ErrInvalidReference, ErrInvalidMinorVersion)
		}
		c.minor = minor
		c.minorSet = true
	}

	return c, nil
}

// Whether the operator requires an explicit upper bound.
//
// The > and >= operators create unbounded ranges that would match any future
// version without limit. Crucible requires these to be paired with an upper
// bound to prevent unintended compatibility with breaking changes.
func requiresUpperBound(op string) bool {
	return op == ">" || op == ">="
}

// Whether the operator provides an implicit upper bound.
//
// Equality (=), inequality (!=), less-than (<, <=), tilde (~), and caret (^)
// operators all constrain the upper version either explicitly or implicitly.
// Hyphen ranges are expanded before validation, so <= appears directly.
func hasUpperBound(op string) bool {
	switch op {
	case "=", "!=", "<", "<=", "~", "^":
		return true
	}
	return false
}

// Validates that a constraint group has proper upper bounds.
//
// Crucible requires all version ranges to have explicit upper bounds to prevent
// unintended compatibility with future major versions. A constraint like
// ">=1.0.0" alone is invalid because it would match version 2.0.0 and beyond,
// which may contain breaking changes.
func validateConstraintGroup(g constraintGroup) error {
	needsUpper := false
	hasUpper := false

	for _, c := range g.constraints {
		if requiresUpperBound(c.operator) {
			needsUpper = true
		}
		if hasUpperBound(c.operator) {
			hasUpper = true
		}
	}

	if needsUpper && !hasUpper {
		return wrap(ErrInvalidReference, ErrMissingUpperBound)
	}

	return nil
}
