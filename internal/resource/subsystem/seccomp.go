package subsystem

import (
	"context"
	"slices"
	"strconv"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Implements [Subsystem] for [manifest.DomainSeccomp].
//
// Accumulates seccomp rules into an internal allow-list. seccomp grants are
// purely additive; duplicate rules are ignored, no conflicts are possible.
type SeccompSubsystem struct {
	rules []seccomp
}

// Feeds a seccomp grant into the subsystem.
//
// Validates the syscall name and argument filters. Bracket lists in args are
// expanded into the cartesian product, producing one rule per combination.
// Each expanded rule is checked for duplicates; only new rules are
// accumulated and returned as normalized grants.
func (s *SeccompSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainSeccomp, "unexpected seccomp domain %q", domain)

	rules, err := parseSeccomp(input.Expr, input.Args)
	if err != nil {
		return nil, err
	}

	grants := s.mergeRules(domain, rules)
	if len(grants) == 0 {
		return nil, nil
	}
	return grants, nil
}

// Deduplicates rules against the accumulated model and returns grants for
// newly added rules.
func (s *SeccompSubsystem) mergeRules(domain Domain, rules []seccomp) []manifest.Grant {
	var grants []manifest.Grant
	for _, r := range rules {
		if s.hasRule(r) {
			continue
		}
		s.rules = append(s.rules, r)
		grants = append(grants, manifest.Grant{
			Subsystem: string(domain),
			Expr:      r.Syscall,
			Args:      serializeSeccompArgs(r.Args),
		})
	}
	return grants
}

// Reports whether the exact rule already exists in the accumulated model.
func (s *SeccompSubsystem) hasRule(r seccomp) bool {
	for _, e := range s.rules {
		if e.Syscall == r.Syscall && slices.Equal(e.Args, r.Args) {
			return true
		}
	}
	return false
}

// Serializes seccomp args back to their text form.
//
// Each arg becomes "<position> <op> <val>" or "<position> <op> <val> <mask>"
// for masked_eq. This is the inverse of parseSeccompArgScalar.
func serializeSeccompArgs(args []seccompArg) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, len(args))
	for i, a := range args {
		s := strconv.FormatUint(uint64(a.Arg), 10) + " " + string(a.Op) + " " + strconv.FormatUint(a.Val, 10)
		if a.Op == seccompOpMaskedEq {
			s += " " + strconv.FormatUint(a.Mask, 10)
		}
		out[i] = s
	}
	return out
}
