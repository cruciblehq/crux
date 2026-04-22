package cap

import (
	"strings"

	"github.com/cruciblehq/crex"

	"github.com/cruciblehq/crux/internal/subsystem/shared"
)

// Parses a capability rule string into a one-rule Model delta.
//
// The rule has the form "[mode] name" where mode selects which capability
// sets to populate and name is a Linux capability without the CAP_ prefix.
// When mode is omitted, ModeFull is assumed. Returns an error if the rule
// is malformed, or either the mode or capability are unknown.
func Parse(rule string) (*Model, error) {
	mode, name, err := ParseRule(rule)
	if err != nil {
		return nil, err
	}
	delta := NewModel()
	switch mode {
	case ModeFull:
		delta.grantFull(name)
	case ModeEffective:
		delta.grantEffective(name)
	case ModeInheritable:
		delta.grantInheritable(name)
	case ModePermitted:
		delta.grantPermitted(name)
	case ModeBound:
		delta.grantBound(name)
	default:
		return nil, crex.Wrapf(ErrInvalidRule, "unknown capability mode %q", mode)
	}
	return delta, nil
}

// Parses a capability rule string into a mode and capability name.
//
// The rule has the form "[mode] name" where mode selects which capability
// sets to populate and name is a Linux capability without the CAP_ prefix.
// When mode is omitted, ModeFull is assumed.
func ParseRule(rule string) (Mode, string, error) {
	fields := strings.Fields(rule)
	if len(fields) == 0 {
		return "", "", crex.Wrapf(ErrInvalidRule, "capability rule is empty")
	}
	if len(fields) > 2 {
		return "", "", crex.Wrapf(ErrInvalidRule, "unexpected token %q", fields[2])
	}

	mode := ModeFull
	name := fields[0]

	if len(fields) > 1 {
		m, err := ParseMode(fields[0])
		if err != nil {
			return "", "", err
		}
		mode = m
		name = fields[1]
	}

	if _, err := shared.ParseCap(name); err != nil {
		return "", "", crex.Wrapf(ErrInvalidRule, "unknown capability %q", name)
	}
	return mode, name, nil
}
