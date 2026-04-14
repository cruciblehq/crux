package subsystem

import (
	"strings"

	"github.com/cruciblehq/crex"
)

// Parses a caps expression "<verb> <name>" into a [caps] config.
//
// The verb determines which kernel capability sets are populated. Each verb
// expands to a predefined combination of sets via the corresponding Grant
// method on [caps]. The capability name must be an uppercase identifier
// matching the kernel ABI (e.g., NET_RAW, SYS_PTRACE).
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
	if err := validateCapName(name); err != nil {
		return caps{}, err
	}

	var c caps
	c.grant(verb, name)
	return c, nil
}
