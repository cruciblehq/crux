package rlimit

import (
	"math"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

// Implementation for POSIX resource limits.
//
// Holds a pointer to the OCI rlimits slice header of the unified spec,
// wired in at construction time. The shared baseline pre-populates the
// slice with one entry per known resource at soft=hard=0; Build updates
// matching entries in place rather than appending duplicates.
type Subsystem struct {
	spec     *[]specs.POSIXRlimit
	declared map[string]specs.POSIXRlimit
}

// Returns a Subsystem wired to mutate the rlimits slice header.
func New(spec *[]specs.POSIXRlimit) *Subsystem {
	return &Subsystem{spec: spec, declared: make(map[string]specs.POSIXRlimit)}
}

// Returns the rlimit subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameRlimit
}

// Applies a parsed grant to the wired-in section.
//
// The grant has the form ".rlimit NAME SOFT [HARD]". When HARD is omitted
// it defaults to SOFT.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	e, err := parse(g)
	if err != nil {
		return err
	}
	return s.apply(e)
}

// Validates the grant's structural shape against what the rlimit subsystem accepts.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Newf(ErrInvalidGrant, "unexpected where clause in rlimit expression")
	}
	if len(g.Kwargs) != 0 {
		return crex.Newf(ErrInvalidGrant, "unexpected keyword arguments in rlimit expression")
	}
	if len(g.Args) < 2 || len(g.Args) > 3 {
		return crex.Newf(ErrInvalidGrant, "wrong number of arguments in rlimit expression")
	}
	return nil
}

// Extracts and validates the grant's content into a POSIX rlimit entry.
//
// Validates that the resource name is known, that the soft and hard limits are
// valid integers, and that the soft limit does not exceed the hard limit.
func parse(g *agl.Model) (specs.POSIXRlimit, error) {
	nameArg := g.Args[argResource]
	if nameArg.Type != agl.ArgName {
		return specs.POSIXRlimit{}, crex.Newf(ErrInvalidGrant, "expected name as resource in rlimit expression")
	}
	if _, ok := knownResources[nameArg.Value]; !ok {
		return specs.POSIXRlimit{}, crex.Newf(ErrInvalidGrant, "unknown resource %q", nameArg.Value)
	}
	soft, err := parseLimit(g.Args[argSoft], labelSoft)
	if err != nil {
		return specs.POSIXRlimit{}, err
	}
	hard := soft
	if len(g.Args) > argHard {
		hard, err = parseLimit(g.Args[argHard], labelHard)
		if err != nil {
			return specs.POSIXRlimit{}, err
		}
	}
	if soft > hard {
		return specs.POSIXRlimit{}, crex.Newf(ErrInvalidGrant, "soft limit %d exceeds hard limit %d in rlimit expression", soft, hard)
	}
	full := rlimitNamePrefix + strings.ToUpper(nameArg.Value)
	return specs.POSIXRlimit{Type: full, Soft: soft, Hard: hard}, nil
}

// Writes a parsed entry to the wired-in slice.
//
// On first declaration, an existing baseline entry is updated in place;
// otherwise the entry is appended.
func (s *Subsystem) apply(e specs.POSIXRlimit) error {
	if existing, ok := s.declared[e.Type]; ok {
		if existing.Soft == e.Soft && existing.Hard == e.Hard {
			return nil
		}
		return crex.Newf(ErrInvalidGrant, "%s already declared with different limits", e.Type)
	}
	for i, existing := range *s.spec {
		if existing.Type != e.Type {
			continue
		}
		(*s.spec)[i] = e
		s.declared[e.Type] = e
		return nil
	}
	*s.spec = append(*s.spec, e)
	s.declared[e.Type] = e
	return nil
}

// Parses an integer or "unlimited" argument into an unsigned 64-bit limit.
//
// The keyword "unlimited" maps to RLIM_INFINITY (math.MaxUint64). Integer
// literals are parsed directly.
func parseLimit(a agl.Arg, label string) (uint64, error) {
	if a.Type == agl.ArgName && a.Value == limitUnlimited {
		return math.MaxUint64, nil
	}
	if a.Type != agl.ArgInt {
		return 0, crex.Newf(ErrInvalidGrant, "expected integer or %q as %s limit in rlimit expression", limitUnlimited, label)
	}
	v, err := strconv.ParseUint(a.Value, limitBase, limitBits)
	if err != nil {
		return 0, crex.Newf(ErrInvalidGrant, "invalid %s limit in rlimit expression", label)
	}
	return v, nil
}
