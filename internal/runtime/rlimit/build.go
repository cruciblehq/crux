package rlimit

import (
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

const rlimitNamePrefix = "RLIMIT_"

// All known POSIX resource short names accepted in .rlimit grants.
//
// The kernel exposes these via the RLIMIT_<UPPER> constants. The grant language
// uses the lowercase short form (e.g. "nofile" rather than "RLIMIT_NOFILE").
var knownResources = map[string]struct{}{
	"as": {}, "core": {}, "cpu": {}, "data": {}, "fsize": {},
	"locks": {}, "memlock": {}, "msgqueue": {}, "nice": {},
	"nofile": {}, "nproc": {}, "rss": {}, "rtprio": {},
	"rttime": {}, "sigpending": {}, "stack": {},
}

// Implementation for POSIX resource limits.
//
// Holds a pointer to the OCI rlimits slice header of the unified spec,
// wired in at construction time. The shared baseline pre-populates the
// slice with one entry per known resource at soft=hard=0; Build updates
// matching entries in place rather than appending duplicates.
type Subsystem struct {
	rlimits *[]specs.POSIXRlimit
}

// Returns a Subsystem wired to mutate the rlimits slice header.
func New(rlimits *[]specs.POSIXRlimit) *Subsystem {
	return &Subsystem{rlimits: rlimits}
}

// Returns the rlimit subsystem identifier.
func (s *Subsystem) Name() shared.Name {
	return shared.NameRlimit
}

// Applies a parsed grant to the wired-in section.
//
// The grant has the form ".rlimit NAME SOFT [HARD]". When HARD is omitted
// it defaults to SOFT.
func (s *Subsystem) Build(g grant.Grant) error {
	if err := check(&g); err != nil {
		return err
	}
	e, err := parse(&g)
	if err != nil {
		return err
	}
	return apply(s.rlimits, e)
}

// Folds the rlimit section of src into the wired-in section.
//
// Conflicting entries (same Type, different soft/hard) abort with an error
// wrapping ErrConflict. A missing OCI subtree is a no-op.
func (s *Subsystem) Merge(src shared.Spec) error {
	if src.OCI == nil || src.OCI.Process == nil {
		return nil
	}
	for _, l := range src.OCI.Process.Rlimits {
		if err := apply(s.rlimits, l); err != nil {
			return err
		}
	}
	return nil
}

// Validates the grant's structural shape against what the rlimit subsystem accepts.
func check(g *grant.Grant) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in rlimit expression")
	}
	if len(g.Kwargs) != 0 {
		return crex.Wrapf(ErrInvalidGrant, "unexpected keyword arguments in rlimit expression")
	}
	if len(g.Args) < 2 || len(g.Args) > 3 {
		return crex.Wrapf(ErrInvalidGrant, "wrong number of arguments in rlimit expression")
	}
	return nil
}

// Extracts and validates the grant's content into a POSIX rlimit entry.
//
// Validates that the resource name is known, that the soft and hard limits are
// valid integers, and that the soft limit does not exceed the hard limit.
func parse(g *grant.Grant) (specs.POSIXRlimit, error) {
	nameArg := g.Args[0]
	if nameArg.Type != grant.ArgName {
		return specs.POSIXRlimit{}, crex.Wrapf(ErrInvalidGrant, "expected name as resource in rlimit expression")
	}
	if _, ok := knownResources[nameArg.Value]; !ok {
		return specs.POSIXRlimit{}, crex.Wrapf(ErrInvalidGrant, "unknown resource %q", nameArg.Value)
	}
	soft, err := parseLimit(g.Args[1], "soft")
	if err != nil {
		return specs.POSIXRlimit{}, err
	}
	hard := soft
	if len(g.Args) == 3 {
		hard, err = parseLimit(g.Args[2], "hard")
		if err != nil {
			return specs.POSIXRlimit{}, err
		}
	}
	if soft > hard {
		return specs.POSIXRlimit{}, crex.Wrapf(ErrInvalidGrant, "soft limit %d exceeds hard limit %d in rlimit expression", soft, hard)
	}
	full := rlimitNamePrefix + strings.ToUpper(nameArg.Value)
	return specs.POSIXRlimit{Type: full, Soft: soft, Hard: hard}, nil
}

// Applies a parsed entry to the wired-in slice.
//
// If an entry of the same Type already exists, it is updated in place when
// its current values are still at the deny-all default (0/0) or equal to
// the incoming values. A non-zero existing entry that differs from the
// incoming values is a conflict.
func apply(rlimits *[]specs.POSIXRlimit, e specs.POSIXRlimit) error {
	for i, existing := range *rlimits {
		if existing.Type != e.Type {
			continue
		}
		if existing.Soft == e.Soft && existing.Hard == e.Hard {
			return nil
		}
		if existing.Soft == 0 && existing.Hard == 0 {
			(*rlimits)[i] = e
			return nil
		}
		return crex.Wrapf(ErrConflict, "resource %q already set to %d/%d", e.Type, existing.Soft, existing.Hard)
	}
	*rlimits = append(*rlimits, e)
	return nil
}

// Parses an integer-typed argument into an unsigned 64-bit limit.
func parseLimit(a grant.Arg, label string) (uint64, error) {
	if a.Type != grant.ArgInt {
		return 0, crex.Wrapf(ErrInvalidGrant, "expected integer as %s limit in rlimit expression", label)
	}
	v, err := strconv.ParseUint(a.Value, 0, 64)
	if err != nil {
		return 0, crex.Wrapf(ErrInvalidGrant, "invalid %s limit in rlimit expression", label)
	}
	return v, nil
}
