package subsystem

import "github.com/cruciblehq/crex"

// Comparison operator for seccomp argument filters.
type seccompOp string

const (
	seccompOpEqual    seccompOp = "eq"        // Exact match.
	seccompOpNotEqual seccompOp = "ne"        // Not equal.
	seccompOpGreater  seccompOp = "gt"        // Greater than.
	seccompOpGreaterE seccompOp = "ge"        // Greater than or equal.
	seccompOpLess     seccompOp = "lt"        // Less than.
	seccompOpLessE    seccompOp = "le"        // Less than or equal.
	seccompOpMaskedEq seccompOp = "masked_eq" // (arg & Mask) == Val.
)

// Converts a string to a seccompOp, returning an error for unknown operators.
func parseSeccompOp(s string) (seccompOp, error) {
	op := seccompOp(s)
	switch op {
	case seccompOpEqual, seccompOpNotEqual,
		seccompOpGreater, seccompOpGreaterE,
		seccompOpLess, seccompOpLessE,
		seccompOpMaskedEq:
		return op, nil
	default:
		return "", crex.Wrapf(ErrSeccompArgFilter, "unknown op %q", s)
	}
}
