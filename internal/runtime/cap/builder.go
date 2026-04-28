package cap

import (
	"slices"

	"github.com/cruciblehq/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

// Compiles capability grants into an OCI [specs.LinuxCapabilities].
//
// Owns the accumulated capability sets and deduplicates grants across multiple
// Build calls within a single compilation session. Grants are purely additive
// and idempotent. The zero value of a Builder represents the strictest possible
// policy (no capability is allowed in any kernel set).
type Builder struct {
	caps specs.LinuxCapabilities // Accumulated capability sets.
}

// Returns a new Builder with no accumulated capabilities.
func NewBuilder() *Builder {
	return &Builder{}
}

// Applies a parsed grant to the accumulated capability sets.
//
// The grant has the form ".cap NAME [MODE]" where NAME is a Linux capability
// without the CAP_ prefix and MODE optionally selects which kernel sets to
// populate. When MODE is omitted, ModeFull is used. The grant accepts no
// keyword arguments and no where clause.
func (b *Builder) Build(g *grant.Grant) error {
	if err := b.check(g); err != nil {
		return err
	}
	name, mode, err := b.parse(g)
	if err != nil {
		return err
	}
	return b.apply(name, mode)
}

// Validates the grant's structural shape against what the cap subsystem accepts.
func (b *Builder) check(g *grant.Grant) error {
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
// The capability name is checked against the shared catalog and the mode is
// resolved against the cap-local mode set, defaulting to ModeFull when absent.
func (b *Builder) parse(g *grant.Grant) (shared.Cap, Mode, error) {
	nameArg := g.Args[0]
	if nameArg.Type != grant.ArgName {
		return "", "", crex.Wrapf(ErrInvalidGrant, "expected name as capability in cap expression, found %s instead", nameArg)
	}
	cap, err := shared.ParseCap(nameArg.Value)
	if err != nil {
		return "", "", crex.Wrapf(ErrInvalidGrant, "unknown capability %q in cap expression", nameArg.Value)
	}
	mode := ModeFull
	if len(g.Args) == 2 {
		modeArg := g.Args[1]
		if modeArg.Type != grant.ArgName {
			return "", "", crex.Wrapf(ErrInvalidGrant, "expected name as mode in cap expression, found %s instead", modeArg)
		}
		m, err := ParseMode(modeArg.Value)
		if err != nil {
			return "", "", err
		}
		mode = m
	}
	return cap, mode, nil
}

// Applies a parsed capability and mode to the accumulated capability sets.
//
// Each mode determines which of the five kernel sets receive the capability,
// stored with the canonical CAP_ prefix.
func (b *Builder) apply(cap shared.Cap, mode Mode) error {
	normalized := shared.NormalizeCap(cap)
	switch mode {
	case ModeFull:
		addCap(&b.caps.Effective, normalized)
		addCap(&b.caps.Permitted, normalized)
		addCap(&b.caps.Inheritable, normalized)
		addCap(&b.caps.Bounding, normalized)
		addCap(&b.caps.Ambient, normalized)
	case ModeEffective:
		addCap(&b.caps.Effective, normalized)
		addCap(&b.caps.Permitted, normalized)
		addCap(&b.caps.Bounding, normalized)
	case ModeInheritable:
		addCap(&b.caps.Permitted, normalized)
		addCap(&b.caps.Inheritable, normalized)
		addCap(&b.caps.Ambient, normalized)
		addCap(&b.caps.Bounding, normalized)
	case ModePermitted:
		addCap(&b.caps.Permitted, normalized)
		addCap(&b.caps.Bounding, normalized)
	case ModeBound:
		addCap(&b.caps.Bounding, normalized)
	}
	return nil
}

// Returns a deep copy of the accumulated capability sets.
//
// The zero state (no grants applied) returns a populated value with all
// five sets empty, denoting the strictest possible policy. Callers must
// not interpret a returned value as "no opinion".
func (b *Builder) Spec() *specs.LinuxCapabilities {
	return &specs.LinuxCapabilities{
		Effective:   slices.Clone(b.caps.Effective),
		Permitted:   slices.Clone(b.caps.Permitted),
		Inheritable: slices.Clone(b.caps.Inheritable),
		Bounding:    slices.Clone(b.caps.Bounding),
		Ambient:     slices.Clone(b.caps.Ambient),
	}
}

// Incorporates capabilities from another set into the accumulated state.
//
// Each kernel set is unioned with the corresponding set in other. Missing
// or nil sets are no-ops.
func (b *Builder) Merge(other *specs.LinuxCapabilities) error {
	if other == nil {
		return nil
	}
	addCaps(&b.caps.Effective, other.Effective)
	addCaps(&b.caps.Permitted, other.Permitted)
	addCaps(&b.caps.Inheritable, other.Inheritable)
	addCaps(&b.caps.Bounding, other.Bounding)
	addCaps(&b.caps.Ambient, other.Ambient)
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
