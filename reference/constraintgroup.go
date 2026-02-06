package reference

import "strings"

// Represents constraints joined by AND (space-separated). Multiple groups are
// joined by OR (|| operator) in a VersionConstraint.
//
// A constraint group must contain at least one constraint.
type constraintGroup struct {
	constraints []constraint
}

// Whether a version satisfies all constraints in the group (AND logic).
func (g constraintGroup) matches(v *Version) bool {
	for _, c := range g.constraints {
		if !c.matches(v) {
			return false
		}
	}
	return true
}

// Returns a string representation of the constraint group.
func (g constraintGroup) String() string {
	parts := make([]string, len(g.constraints))
	for i, c := range g.constraints {
		parts[i] = c.String()
	}
	return strings.Join(parts, " ")
}
