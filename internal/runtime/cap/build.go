package cap

import (
	"slices"

	"github.com/cruciblehq/crux/internal/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/aegis"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

// Subsystem implementation for Linux capabilities.
//
// Holds a pointer to the OCI capabilities section of the unified spec, wired
// in at construction time. Build and Merge mutate that section in place.
type Subsystem struct {
	caps *specs.LinuxCapabilities // Pointer to the spec's capabilities section.
}

// Returns a Subsystem wired to mutate caps.
func New(caps *specs.LinuxCapabilities) *Subsystem {
	return &Subsystem{caps: caps}
}

// Returns the cap subsystem identifier.
func (s *Subsystem) Name() shared.Name {
	return shared.NameCap
}

// Applies a parsed grant to the wired-in section.
//
// The grant has the form ".cap NAME [MODE]" where NAME is a Linux capability
// without the CAP_ prefix and MODE optionally selects which kernel sets to
// populate. When MODE is omitted, ModeFull is used. The grant accepts no
// keyword arguments and no where clause.
func (s *Subsystem) Build(g *aegis.Model) error {
	if err := check(g); err != nil {
		return err
	}
	name, mode, err := parse(g)
	if err != nil {
		return err
	}
	return apply(s.caps, name, mode)
}

// Folds the cap section of src into the wired-in section.
//
// Each kernel set in src.OCI.Process.Capabilities is unioned with the
// corresponding set in the receiver's section. A missing OCI subtree or
// nil capabilities pointer is a no-op.
func (s *Subsystem) Merge(src shared.Spec) error {
	if src.OCI == nil || src.OCI.Process == nil || src.OCI.Process.Capabilities == nil {
		return nil
	}
	caps := src.OCI.Process.Capabilities
	addCaps(&s.caps.Effective, caps.Effective)
	addCaps(&s.caps.Permitted, caps.Permitted)
	addCaps(&s.caps.Inheritable, caps.Inheritable)
	addCaps(&s.caps.Bounding, caps.Bounding)
	addCaps(&s.caps.Ambient, caps.Ambient)
	return nil
}

// Validates the grant's structural shape against what the cap subsystem accepts.
func check(g *aegis.Model) error {
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
func parse(g *aegis.Model) (shared.Cap, mode, error) {
	nameArg := g.Args[0]
	if nameArg.Type != aegis.ArgName {
		return "", "", crex.Wrapf(ErrInvalidGrant, "expected name as capability in cap expression, found %s instead", nameArg)
	}
	c, err := shared.ParseCap(nameArg.Value)
	if err != nil {
		return "", "", crex.Wrapf(ErrInvalidGrant, "unknown capability %q in cap expression", nameArg.Value)
	}
	mode := modeFull
	if len(g.Args) == 2 {
		modeArg := g.Args[1]
		if modeArg.Type != aegis.ArgName {
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
func apply(caps *specs.LinuxCapabilities, c shared.Cap, mode mode) error {
	normalized := shared.NormalizeCap(c)
	switch mode {
	case modeFull:
		addCap(&caps.Effective, normalized)
		addCap(&caps.Permitted, normalized)
		addCap(&caps.Inheritable, normalized)
		addCap(&caps.Bounding, normalized)
		addCap(&caps.Ambient, normalized)
	case modeEffective:
		addCap(&caps.Effective, normalized)
		addCap(&caps.Permitted, normalized)
		addCap(&caps.Bounding, normalized)
	case modeInheritable:
		addCap(&caps.Permitted, normalized)
		addCap(&caps.Inheritable, normalized)
		addCap(&caps.Ambient, normalized)
		addCap(&caps.Bounding, normalized)
	case modePermitted:
		addCap(&caps.Permitted, normalized)
		addCap(&caps.Bounding, normalized)
	case modeBound:
		addCap(&caps.Bounding, normalized)
	}
	return nil
}

// Grants all capabilities in src to dst, returning true if dst was changed.
func addCaps(dst *[]string, src []string) bool {
	changed := false
	for _, c := range src {
		if addCap(dst, c) {
			changed = true
		}
	}
	return changed
}

// Appends c to dst if not already present, returning true if dst was changed.
func addCap(dst *[]string, c string) bool {
	if slices.Contains(*dst, c) {
		return false
	}
	*dst = append(*dst, c)
	return true
}
