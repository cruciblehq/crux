package macspec

// Accumulated MAC grant spec.
//
// Tracks which LSM hook rules have been granted. MAC grants are purely
// additive. A rule with the same hook name and structurally identical
// where-clause is dropped as a no-op. Multiple rules for the same hook
// with different where-clauses are kept because they are OR'd by the
// enforcer.
type Spec struct {
	Rules []*Allow `json:"rules"` // Granted LSM hook allow rules.
}

// Returns a ready-to-use MAC spec.
func NewSpec() *Spec {
	return &Spec{}
}

// Subsystem-specific rule expression for MAC grants.
//
// Each Allow represents a single allow directive targeting a kernel LSM
// hook, with an optional where-clause filter tree.
type Allow struct {
	Hook  string `json:"hook"`  // Kernel LSM hook name such as "file_open" or "socket_create".
	Where *Expr  `json:"where"` // Where-clause expression tree. Nil when the grant is unconditional.
}

// Merges a rule into the accumulated spec.
//
// Returns true if the rule was not already present. Two rules are
// considered equal when the hook names match and the where-clause
// expression trees are structurally identical.
func (s *Spec) Apply(r *Allow) bool {
	for _, existing := range s.Rules {
		if existing.Hook == r.Hook && exprEqual(existing.Where, r.Where) {
			return false
		}
	}
	s.Rules = append(s.Rules, cloneAllow(r))
	return true
}

// Folds other into the receiver.
//
// Each rule in other is merged via [Spec.Apply] which performs semantic
// deduplication. A nil other is a no-op.
func (s *Spec) MergeSpec(other *Spec) {
	if other == nil {
		return
	}
	for _, a := range other.Rules {
		s.Apply(a)
	}
}

// Returns a deep copy of a, or nil when a is nil.
func cloneAllow(a *Allow) *Allow {
	if a == nil {
		return nil
	}
	return &Allow{Hook: a.Hook, Where: cloneExpr(a.Where)}
}
