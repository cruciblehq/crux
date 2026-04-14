package subsystem

import (
	"regexp"
	"strings"

	"github.com/cruciblehq/crex"
)

// Matches a valid capability name (uppercase letters, digits, and underscores).
var validCapName = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)

// Parses a caps expression "<verb> <name>" into a [caps] config.
func parseCaps(expr string) (caps, error) {
	fields := strings.Fields(expr)
	if len(fields) != 2 {
		return caps{}, crex.Wrapf(ErrGrantExpression, "invalid expression %q", expr)
	}
	verb, err := parseCapVerb(fields[0])
	if err != nil {
		return caps{}, err
	}
	name := fields[1]
	if !validCapName.MatchString(name) {
		return caps{}, crex.Wrapf(ErrGrantExpression, "invalid capability name %q", name)
	}

	var c caps
	switch verb {
	case capVerbGrant:
		c.Grant(name)
	case capVerbEffective:
		c.GrantEffective(name)
	case capVerbInheritable:
		c.GrantInheritable(name)
	case capVerbPermitted:
		c.GrantPermitted(name)
	case capVerbBound:
		c.GrantBound(name)
	}
	return c, nil
}
