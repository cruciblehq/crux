package mac

import "github.com/cruciblehq/crux/crex"

// Accumulated MAC grant spec.
//
// Built up by successive calls to Apply during affordance evaluation.
// Duplicate rules are suppressed so the slice contains only the minimal set
// of distinct hook allows required by the workload.
type Spec struct {
	Rules []*MACAllow `codec:"rules"` // Granted LSM hook allow rules.
}

// Appends r to the spec's rule list if it is not already present.
//
// Returns true if the rule was added (false means a rule with the same hook
// and where clause already existed, and the call was a no-op).
func (s *Spec) Apply(r *MACAllow) bool {
	for _, existing := range s.Rules {
		if existing.Hook == r.Hook && macExprEqual(existing.Where, r.Where) {
			return false
		}
	}
	s.Rules = append(s.Rules, cloneMACAllow(r))
	return true
}

// Validates the MAC spec.
func (s *Spec) Validate() error {
	for i, rule := range s.Rules {
		if rule == nil {
			return crex.Wrapf(ErrInvalidMAC, "nil rule at index %d", i)
		}
		if err := rule.Validate(); err != nil {
			return crex.Wrap(ErrInvalidMAC, err)
		}
	}
	return nil
}
