package cap

import (
	"slices"

	"github.com/cruciblehq/crux/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/capset"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

// Implementation of the Linux capabilities subsystem.
//
// Holds a pointer to the OCI capabilities section of the unified spec, wired
// in at construction time. Build and Merge mutate that section in place.
type Subsystem struct {
	spec *specs.LinuxCapabilities
}

// Returns a Subsystem wired to mutate caps.
func New(spec *specs.LinuxCapabilities) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the cap subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameCap
}

// Returns the deduplication key for a cap grant.
//
// The key is the capability name (args[0]). This means .cap net_admin and
// .cap net_admin full are treated as conflicts: the same capability may not
// appear in more than one grant regardless of mode.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) == 0 {
		return ""
	}
	return g.Args[0].Value
}

// Applies a parsed grant to the wired-in section.
//
// The grant has the form ".cap NAME [MODE]" where NAME is a Linux capability
// (without the CAP_ prefix) and MODE selects which kernel sets to populate.
// When MODE is omitted, ModeFull is used.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	name, mode, err := parse(g)
	if err != nil {
		return err
	}
	return apply(s.spec, name, mode)
}

// Validates the grant's structural shape against what the cap subsystem accepts.
//
// Cap grants have no where clause or kwargs, and one or two args: a capability
// name and optional mode. Returns an error if the grant fails this check.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in cap expression")
	}
	if len(g.Kwargs) != 0 {
		return crex.Wrapf(ErrInvalidGrant, "unexpected keyword arguments in cap expression")
	}
	if len(g.Args) == 0 {
		return crex.Wrapf(ErrInvalidGrant, "missing capability name in cap expression")
	}
	if len(g.Args) > 2 {
		return crex.Wrapf(ErrInvalidGrant, "too many arguments in cap expression")
	}
	return nil
}

// Extracts and validates the grant's capability name and optional mode.
//
// The first argument is the capability name, which must be a valid Linux
// capability (without the CAP_ prefix). The second argument, if present, is
// the mode, which selects which kernel sets to populate. The mode defaults
// to ModeFull if not present. Returns an error if the capability name is
// unknown or if the mode is invalid.
func parse(g *agl.Model) (capset.Cap, mode, error) {
	nameArg := g.Args[0]
	if nameArg.Type != agl.ArgName {
		return "", "", crex.Wrapf(ErrInvalidGrant, "expected name as capability in cap expression, found %s instead", nameArg)
	}
	c, err := capset.Parse(nameArg.Value)
	if err != nil {
		return "", "", crex.Wrapf(ErrInvalidGrant, "unknown capability %q in cap expression", nameArg.Value)
	}
	mode := modeFull
	if len(g.Args) == 2 {
		modeArg := g.Args[1]
		if modeArg.Type != agl.ArgName {
			return "", "", crex.Wrapf(ErrInvalidGrant, "expected name as mode in cap expression, found %s instead", modeArg)
		}
		m, err := parseMode(modeArg.Value)
		if err != nil {
			return "", "", err
		}
		mode = m
	}
	return c, mode, nil
}

// Applies a parsed capability and mode to the wired-in section.
func apply(lc *specs.LinuxCapabilities, c capset.Cap, mode mode) error {
	normalized := capset.Normalize(c)
	switch mode {
	case modeFull:
		addCap(&lc.Effective, normalized)
		addCap(&lc.Permitted, normalized)
		addCap(&lc.Inheritable, normalized)
		addCap(&lc.Bounding, normalized)
		addCap(&lc.Ambient, normalized)
	case modeEffective:
		addCap(&lc.Effective, normalized)
		addCap(&lc.Permitted, normalized)
		addCap(&lc.Bounding, normalized)
	case modeInheritable:
		addCap(&lc.Permitted, normalized)
		addCap(&lc.Inheritable, normalized)
		addCap(&lc.Ambient, normalized)
		addCap(&lc.Bounding, normalized)
	case modePermitted:
		addCap(&lc.Permitted, normalized)
		addCap(&lc.Bounding, normalized)
	case modeBound:
		addCap(&lc.Bounding, normalized)
	}
	return nil
}

// Appends c to dst if not already present, returning true if dst was changed.
func addCap(dst *[]string, c string) bool {
	if slices.Contains(*dst, c) {
		return false
	}
	*dst = append(*dst, c)
	return true
}
