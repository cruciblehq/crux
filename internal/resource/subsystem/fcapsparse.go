package subsystem

import (
	"strings"

	"github.com/cruciblehq/crex"
)

// Parses an fcap expression "<verb> <path> <caps...>" into an [fcap] config.
//
// The verb selects which file capability sets are populated: fcapVerbEffective
// adds each capability to file-permitted and sets the effective bit, while
// fcapVerbInheritable adds to file-inheritable only. The path is the binary's
// absolute path inside the container. One or more capability names follow.
func parseFcap(expr string) (fcap, error) {
	fields := strings.Fields(expr)
	if len(fields) < 3 {
		return fcap{}, crex.Wrapf(ErrGrantExpression, "invalid expression %q", expr)
	}
	verb, err := parseFcapVerb(fields[0])
	if err != nil {
		return fcap{}, err
	}
	path, err := validateContainerPath(fields[1])
	if err != nil {
		return fcap{}, err
	}
	names := fields[2:]
	for _, name := range names {
		if err := validateCapName(name); err != nil {
			return fcap{}, err
		}
	}

	var fc fcap
	fc.Path = path
	for _, name := range names {
		fc.grant(verb, name)
	}
	return fc, nil
}
