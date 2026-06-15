package mac

import "github.com/cruciblehq/crux/crex"

// Single LSM hook allow rule produced by a MAC grant.
//
// Hook names the kernel LSM callback point at which the rule is evaluated.
// Where constrains the allow to invocations matching the predicate tree; a
// nil Where makes the rule unconditional.
type MACAllow struct {
	Hook  string   `codec:"hook"`  // Kernel LSM hook name.
	Where *MACExpr `codec:"where"` // Where-clause expression tree.
}

// Deep-clones a MACAllow.
func cloneMACAllow(a *MACAllow) *MACAllow {
	return &MACAllow{
		Hook:  a.Hook,
		Where: cloneMACExpr(a.Where),
	}
}

// Validates the allow rule.
func (a *MACAllow) Validate() error {
	if a.Hook == "" {
		return crex.Newf(ErrInvalidMACAllow, "hook is empty")
	}
	if a.Where != nil {
		if err := a.Where.Validate(); err != nil {
			return crex.Wrap(ErrInvalidMACAllow, err)
		}
	}
	return nil
}
