package cap

import (
	"strings"

	"github.com/cruciblehq/crex"

	"github.com/cruciblehq/crux/internal/subsystem/shared"
)

// Parses a capability rule string into a Grant.
//
// The rule has the form "[mode] name" where mode selects which capability
// sets to populate and name is a Linux capability without the CAP_ prefix.
// When mode is omitted, "full" is assumed (all five sets).
func Parse(rule string) (*Grant, error) {
	fields := strings.Fields(rule)
	if len(fields) == 0 {
		return nil, crex.Wrapf(ErrInvalidRule, "capability rule is empty")
	}
	if len(fields) > 2 {
		return nil, crex.Wrapf(ErrInvalidRule, "unexpected token %q", fields[2])
	}

	mode := ModeFull
	name := fields[0]

	if len(fields) > 1 {
		m, err := ParseMode(fields[0])
		if err != nil {
			return nil, err
		}
		mode = m
		name = fields[1]
	}

	if name == "" {
		return nil, crex.Wrapf(ErrInvalidRule, "capability name is empty")
	}
	if _, err := shared.ParseCap(name); err != nil {
		return nil, crex.Wrapf(ErrInvalidRule, "unknown capability %q", name)
	}
	return &Grant{Mode: mode, Name: name}, nil
}
