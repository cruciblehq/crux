package reference

import "testing"

// Equal operator tests

func TestConstraint_Equal_ExactMatch(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.2.3")) {
		t.Error("=1.2.3 should match 1.2.3")
	}
}

func TestConstraint_Equal_MismatchMajor(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "2.2.3")) {
		t.Error("=1.2.3 should not match 2.2.3")
	}
}

func TestConstraint_Equal_MismatchMinor(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.3.3")) {
		t.Error("=1.2.3 should not match 1.3.3")
	}
}

func TestConstraint_Equal_MismatchPatch(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.2.4")) {
		t.Error("=1.2.3 should not match 1.2.4")
	}
}

func TestConstraint_Equal_MajorOnly_Matches(t *testing.T) {
	c := constraint{operator: "=", major: 1}
	if !c.matches(mustParseVersion(t, "1.0.0")) {
		t.Error("=1 should match 1.0.0")
	}
}

func TestConstraint_Equal_MajorOnly_MatchesAnyMinor(t *testing.T) {
	c := constraint{operator: "=", major: 1}
	if !c.matches(mustParseVersion(t, "1.5.0")) {
		t.Error("=1 should match 1.5.0")
	}
}

func TestConstraint_Equal_MajorOnly_MatchesAnyPatch(t *testing.T) {
	c := constraint{operator: "=", major: 1}
	if !c.matches(mustParseVersion(t, "1.5.9")) {
		t.Error("=1 should match 1.5.9")
	}
}

func TestConstraint_Equal_MajorOnly_Mismatch(t *testing.T) {
	c := constraint{operator: "=", major: 1}
	if c.matches(mustParseVersion(t, "2.0.0")) {
		t.Error("=1 should not match 2.0.0")
	}
}

func TestConstraint_Equal_MajorMinor_Matches(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, minorSet: true}
	if !c.matches(mustParseVersion(t, "1.2.0")) {
		t.Error("=1.2 should match 1.2.0")
	}
}

func TestConstraint_Equal_MajorMinor_MatchesAnyPatch(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, minorSet: true}
	if !c.matches(mustParseVersion(t, "1.2.9")) {
		t.Error("=1.2 should match 1.2.9")
	}
}

func TestConstraint_Equal_MajorMinor_Mismatch(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, minorSet: true}
	if c.matches(mustParseVersion(t, "1.3.0")) {
		t.Error("=1.2 should not match 1.3.0")
	}
}

// NotEqual operator tests

func TestConstraint_NotEqual_Matches(t *testing.T) {
	c := constraint{operator: "!=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.2.4")) {
		t.Error("!=1.2.3 should match 1.2.4")
	}
}

func TestConstraint_NotEqual_Mismatch(t *testing.T) {
	c := constraint{operator: "!=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.2.3")) {
		t.Error("!=1.2.3 should not match 1.2.3")
	}
}

func TestConstraint_NotEqual_DifferentMajor(t *testing.T) {
	c := constraint{operator: "!=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "2.0.0")) {
		t.Error("!=1.2.3 should match 2.0.0")
	}
}

// GreaterThan operator tests

func TestConstraint_GreaterThan_GreaterMajor(t *testing.T) {
	c := constraint{operator: ">", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "2.0.0")) {
		t.Error(">1.0.0 should match 2.0.0")
	}
}

func TestConstraint_GreaterThan_GreaterMinor(t *testing.T) {
	c := constraint{operator: ">", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.1.0")) {
		t.Error(">1.0.0 should match 1.1.0")
	}
}

func TestConstraint_GreaterThan_GreaterPatch(t *testing.T) {
	c := constraint{operator: ">", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.0.1")) {
		t.Error(">1.0.0 should match 1.0.1")
	}
}

func TestConstraint_GreaterThan_Equal(t *testing.T) {
	c := constraint{operator: ">", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.0.0")) {
		t.Error(">1.0.0 should not match 1.0.0")
	}
}

func TestConstraint_GreaterThan_Less(t *testing.T) {
	c := constraint{operator: ">", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "0.9.9")) {
		t.Error(">1.0.0 should not match 0.9.9")
	}
}

// GreaterThanOrEqual operator tests (standard semver behavior)

func TestConstraint_GreaterThanOrEqual_Equal(t *testing.T) {
	c := constraint{operator: ">=", major: 1, minor: 2, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.2.0")) {
		t.Error(">=1.2.0 should match 1.2.0")
	}
}

func TestConstraint_GreaterThanOrEqual_GreaterMinor(t *testing.T) {
	c := constraint{operator: ">=", major: 1, minor: 2, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.3.0")) {
		t.Error(">=1.2.0 should match 1.3.0")
	}
}

func TestConstraint_GreaterThanOrEqual_GreaterMajor(t *testing.T) {
	c := constraint{operator: ">=", major: 1, minor: 2, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "2.0.0")) {
		t.Error(">=1.2.0 should match 2.0.0")
	}
}

func TestConstraint_GreaterThanOrEqual_Less(t *testing.T) {
	c := constraint{operator: ">=", major: 1, minor: 2, patch: 0, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.1.0")) {
		t.Error(">=1.2.0 should not match 1.1.0")
	}
}

// LessThan operator tests

func TestConstraint_LessThan_LessMajor(t *testing.T) {
	c := constraint{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.0.0")) {
		t.Error("<2.0.0 should match 1.0.0")
	}
}

func TestConstraint_LessThan_LessMinor(t *testing.T) {
	c := constraint{operator: "<", major: 1, minor: 5, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.4.0")) {
		t.Error("<1.5.0 should match 1.4.0")
	}
}

func TestConstraint_LessThan_LessPatch(t *testing.T) {
	c := constraint{operator: "<", major: 1, minor: 0, patch: 5, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.0.4")) {
		t.Error("<1.0.5 should match 1.0.4")
	}
}

func TestConstraint_LessThan_Equal(t *testing.T) {
	c := constraint{operator: "<", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.0.0")) {
		t.Error("<1.0.0 should not match 1.0.0")
	}
}

func TestConstraint_LessThan_Greater(t *testing.T) {
	c := constraint{operator: "<", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.0.1")) {
		t.Error("<1.0.0 should not match 1.0.1")
	}
}

// LessThanOrEqual operator tests

func TestConstraint_LessThanOrEqual_Equal(t *testing.T) {
	c := constraint{operator: "<=", major: 1, minor: 2, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.2.0")) {
		t.Error("<=1.2.0 should match 1.2.0")
	}
}

func TestConstraint_LessThanOrEqual_LessMajor(t *testing.T) {
	c := constraint{operator: "<=", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.0.0")) {
		t.Error("<=2.0.0 should match 1.0.0")
	}
}

func TestConstraint_LessThanOrEqual_LessMinor(t *testing.T) {
	c := constraint{operator: "<=", major: 1, minor: 5, patch: 0, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.4.0")) {
		t.Error("<=1.5.0 should match 1.4.0")
	}
}

func TestConstraint_LessThanOrEqual_Greater(t *testing.T) {
	c := constraint{operator: "<=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.0.1")) {
		t.Error("<=1.0.0 should not match 1.0.1")
	}
}

// Tilde operator tests

func TestConstraint_Tilde_MajorOnly_Matches(t *testing.T) {
	c := constraint{operator: "~", major: 1}
	if !c.matchTilde(mustParseVersion(t, "1.0.0")) {
		t.Error("~1 should match 1.0.0")
	}
}

func TestConstraint_Tilde_MajorOnly_MatchesHighMinor(t *testing.T) {
	c := constraint{operator: "~", major: 1}
	if !c.matchTilde(mustParseVersion(t, "1.5.9")) {
		t.Error("~1 should match 1.5.9")
	}
}

func TestConstraint_Tilde_MajorOnly_RejectsDifferentMajor(t *testing.T) {
	c := constraint{operator: "~", major: 1}
	if c.matchTilde(mustParseVersion(t, "2.0.0")) {
		t.Error("~1 should not match 2.0.0")
	}
}

func TestConstraint_Tilde_MajorOnly_RejectsLowerMajor(t *testing.T) {
	c := constraint{operator: "~", major: 1}
	if c.matchTilde(mustParseVersion(t, "0.9.9")) {
		t.Error("~1 should not match 0.9.9")
	}
}

func TestConstraint_Tilde_MajorMinor_Matches(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, minorSet: true}
	if !c.matchTilde(mustParseVersion(t, "1.2.0")) {
		t.Error("~1.2 should match 1.2.0")
	}
}

func TestConstraint_Tilde_MajorMinor_MatchesAnyPatch(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, minorSet: true}
	if !c.matchTilde(mustParseVersion(t, "1.2.9")) {
		t.Error("~1.2 should match 1.2.9")
	}
}

func TestConstraint_Tilde_MajorMinor_RejectsHigherMinor(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, minorSet: true}
	if c.matchTilde(mustParseVersion(t, "1.3.0")) {
		t.Error("~1.2 should not match 1.3.0")
	}
}

func TestConstraint_Tilde_MajorMinor_RejectsLowerMinor(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, minorSet: true}
	if c.matchTilde(mustParseVersion(t, "1.1.9")) {
		t.Error("~1.2 should not match 1.1.9")
	}
}

func TestConstraint_Tilde_Full_Matches(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matchTilde(mustParseVersion(t, "1.2.3")) {
		t.Error("~1.2.3 should match 1.2.3")
	}
}

func TestConstraint_Tilde_Full_MatchesHigherPatch(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matchTilde(mustParseVersion(t, "1.2.9")) {
		t.Error("~1.2.3 should match 1.2.9")
	}
}

func TestConstraint_Tilde_Full_RejectsLowerPatch(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchTilde(mustParseVersion(t, "1.2.2")) {
		t.Error("~1.2.3 should not match 1.2.2")
	}
}

func TestConstraint_Tilde_Full_RejectsHigherMinor(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchTilde(mustParseVersion(t, "1.3.0")) {
		t.Error("~1.2.3 should not match 1.3.0")
	}
}

// Tilde operator through matches()

func TestConstraint_Matches_Tilde_MajorOnly(t *testing.T) {
	c := constraint{operator: "~", major: 1}
	if !c.matches(mustParseVersion(t, "1.5.9")) {
		t.Error("~1 should match 1.5.9")
	}
}

func TestConstraint_Matches_Tilde_MajorOnly_Rejects(t *testing.T) {
	c := constraint{operator: "~", major: 1}
	if c.matches(mustParseVersion(t, "2.0.0")) {
		t.Error("~1 should not match 2.0.0")
	}
}

func TestConstraint_Matches_Tilde_Full(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.2.5")) {
		t.Error("~1.2.3 should match 1.2.5")
	}
}

func TestConstraint_Matches_Tilde_Full_Rejects(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "1.3.0")) {
		t.Error("~1.2.3 should not match 1.3.0")
	}
}

// Caret operator tests - major non-zero

func TestConstraint_Caret_MajorOnly_Matches(t *testing.T) {
	c := constraint{operator: "^", major: 1}
	if !c.matchCaret(mustParseVersion(t, "1.0.0")) {
		t.Error("^1 should match 1.0.0")
	}
}

func TestConstraint_Caret_MajorOnly_MatchesHighMinor(t *testing.T) {
	c := constraint{operator: "^", major: 1}
	if !c.matchCaret(mustParseVersion(t, "1.9.9")) {
		t.Error("^1 should match 1.9.9")
	}
}

func TestConstraint_Caret_MajorOnly_RejectsHigherMajor(t *testing.T) {
	c := constraint{operator: "^", major: 1}
	if c.matchCaret(mustParseVersion(t, "2.0.0")) {
		t.Error("^1 should not match 2.0.0")
	}
}

func TestConstraint_Caret_MajorOnly_RejectsLowerMajor(t *testing.T) {
	c := constraint{operator: "^", major: 1}
	if c.matchCaret(mustParseVersion(t, "0.9.9")) {
		t.Error("^1 should not match 0.9.9")
	}
}

func TestConstraint_Caret_MajorMinor_Matches(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, minorSet: true}
	if !c.matchCaret(mustParseVersion(t, "1.2.0")) {
		t.Error("^1.2 should match 1.2.0")
	}
}

func TestConstraint_Caret_MajorMinor_MatchesHigherMinor(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, minorSet: true}
	if !c.matchCaret(mustParseVersion(t, "1.3.0")) {
		t.Error("^1.2 should match 1.3.0")
	}
}

func TestConstraint_Caret_MajorMinor_MatchesHighMinor(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, minorSet: true}
	if !c.matchCaret(mustParseVersion(t, "1.9.0")) {
		t.Error("^1.2 should match 1.9.0")
	}
}

func TestConstraint_Caret_MajorMinor_RejectsLowerMinor(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, minorSet: true}
	if c.matchCaret(mustParseVersion(t, "1.1.9")) {
		t.Error("^1.2 should not match 1.1.9")
	}
}

func TestConstraint_Caret_MajorMinor_RejectsHigherMajor(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, minorSet: true}
	if c.matchCaret(mustParseVersion(t, "2.0.0")) {
		t.Error("^1.2 should not match 2.0.0")
	}
}

func TestConstraint_Caret_Full_Matches(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matchCaret(mustParseVersion(t, "1.2.3")) {
		t.Error("^1.2.3 should match 1.2.3")
	}
}

func TestConstraint_Caret_Full_MatchesHigherPatch(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matchCaret(mustParseVersion(t, "1.2.9")) {
		t.Error("^1.2.3 should match 1.2.9")
	}
}

func TestConstraint_Caret_Full_MatchesHigherMinor(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matchCaret(mustParseVersion(t, "1.3.0")) {
		t.Error("^1.2.3 should match 1.3.0")
	}
}

func TestConstraint_Caret_Full_RejectsLowerPatch(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "1.2.2")) {
		t.Error("^1.2.3 should not match 1.2.2")
	}
}

func TestConstraint_Caret_Full_RejectsLowerMinor(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "1.1.0")) {
		t.Error("^1.2.3 should not match 1.1.0")
	}
}

func TestConstraint_Caret_Full_RejectsHigherMajor(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "2.0.0")) {
		t.Error("^1.2.3 should not match 2.0.0")
	}
}

// Caret operator tests - major zero

func TestConstraint_Caret_Zero_MajorOnly_Matches(t *testing.T) {
	c := constraint{operator: "^", major: 0}
	if !c.matchCaret(mustParseVersion(t, "0.0.0")) {
		t.Error("^0 should match 0.0.0")
	}
}

func TestConstraint_Caret_Zero_MajorOnly_MatchesAny(t *testing.T) {
	c := constraint{operator: "^", major: 0}
	if !c.matchCaret(mustParseVersion(t, "0.5.9")) {
		t.Error("^0 should match 0.5.9")
	}
}

func TestConstraint_Caret_Zero_MajorOnly_RejectsHigherMajor(t *testing.T) {
	c := constraint{operator: "^", major: 0}
	if c.matchCaret(mustParseVersion(t, "1.0.0")) {
		t.Error("^0 should not match 1.0.0")
	}
}

func TestConstraint_Caret_ZeroZero_Matches(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, minorSet: true}
	if !c.matchCaret(mustParseVersion(t, "0.0.0")) {
		t.Error("^0.0 should match 0.0.0")
	}
}

func TestConstraint_Caret_ZeroZero_MatchesAnyPatch(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, minorSet: true}
	if !c.matchCaret(mustParseVersion(t, "0.0.9")) {
		t.Error("^0.0 should match 0.0.9")
	}
}

func TestConstraint_Caret_ZeroZero_RejectsHigherMinor(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, minorSet: true}
	if c.matchCaret(mustParseVersion(t, "0.1.0")) {
		t.Error("^0.0 should not match 0.1.0")
	}
}

func TestConstraint_Caret_ZeroZeroThree_MatchesExact(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, patch: 3, minorSet: true, patchSet: true}
	if !c.matchCaret(mustParseVersion(t, "0.0.3")) {
		t.Error("^0.0.3 should match 0.0.3")
	}
}

func TestConstraint_Caret_ZeroZeroThree_RejectsLowerPatch(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "0.0.2")) {
		t.Error("^0.0.3 should not match 0.0.2")
	}
}

func TestConstraint_Caret_ZeroZeroThree_RejectsHigherPatch(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "0.0.4")) {
		t.Error("^0.0.3 should not match 0.0.4")
	}
}

func TestConstraint_Caret_ZeroZeroThree_RejectsHigherMinor(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "0.1.0")) {
		t.Error("^0.0.3 should not match 0.1.0")
	}
}

func TestConstraint_Caret_ZeroTwo_Matches(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, minorSet: true}
	if !c.matchCaret(mustParseVersion(t, "0.2.0")) {
		t.Error("^0.2 should match 0.2.0")
	}
}

func TestConstraint_Caret_ZeroTwo_MatchesAnyPatch(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, minorSet: true}
	if !c.matchCaret(mustParseVersion(t, "0.2.9")) {
		t.Error("^0.2 should match 0.2.9")
	}
}

func TestConstraint_Caret_ZeroTwo_RejectsLowerMinor(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, minorSet: true}
	if c.matchCaret(mustParseVersion(t, "0.1.0")) {
		t.Error("^0.2 should not match 0.1.0")
	}
}

func TestConstraint_Caret_ZeroTwo_RejectsHigherMinor(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, minorSet: true}
	if c.matchCaret(mustParseVersion(t, "0.3.0")) {
		t.Error("^0.2 should not match 0.3.0")
	}
}

func TestConstraint_Caret_ZeroTwoThree_Matches(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matchCaret(mustParseVersion(t, "0.2.3")) {
		t.Error("^0.2.3 should match 0.2.3")
	}
}

func TestConstraint_Caret_ZeroTwoThree_MatchesHigherPatch(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matchCaret(mustParseVersion(t, "0.2.9")) {
		t.Error("^0.2.3 should match 0.2.9")
	}
}

func TestConstraint_Caret_ZeroTwoThree_RejectsLowerPatch(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "0.2.2")) {
		t.Error("^0.2.3 should not match 0.2.2")
	}
}

func TestConstraint_Caret_ZeroTwoThree_RejectsHigherMinor(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "0.3.0")) {
		t.Error("^0.2.3 should not match 0.3.0")
	}
}

func TestConstraint_Caret_ZeroTwoThree_RejectsLowerMinor(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matchCaret(mustParseVersion(t, "0.1.0")) {
		t.Error("^0.2.3 should not match 0.1.0")
	}
}

// Caret operator through matches()

func TestConstraint_Matches_Caret_MajorNonZero(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "1.3.0")) {
		t.Error("^1.2.3 should match 1.3.0")
	}
}

func TestConstraint_Matches_Caret_MajorNonZero_Rejects(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "2.0.0")) {
		t.Error("^1.2.3 should not match 2.0.0")
	}
}

func TestConstraint_Matches_Caret_MajorZero(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "0.2.5")) {
		t.Error("^0.2.3 should match 0.2.5")
	}
}

func TestConstraint_Matches_Caret_MajorZero_Rejects(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "0.3.0")) {
		t.Error("^0.2.3 should not match 0.3.0")
	}
}

func TestConstraint_Matches_Caret_ZeroZero_ExactMatch(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, patch: 3, minorSet: true, patchSet: true}
	if !c.matches(mustParseVersion(t, "0.0.3")) {
		t.Error("^0.0.3 should match 0.0.3")
	}
}

func TestConstraint_Matches_Caret_ZeroZero_RejectsHigher(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 0, patch: 3, minorSet: true, patchSet: true}
	if c.matches(mustParseVersion(t, "0.0.4")) {
		t.Error("^0.0.3 should not match 0.0.4")
	}
}

// Compare tests

func TestConstraint_Compare_EqualFull(t *testing.T) {
	c := constraint{major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if got := c.compare(mustParseVersion(t, "1.2.3")); got != 0 {
		t.Errorf("compare() = %d, want 0", got)
	}
}

func TestConstraint_Compare_EqualMajorOnly(t *testing.T) {
	c := constraint{major: 1}
	if got := c.compare(mustParseVersion(t, "1.5.9")); got != 0 {
		t.Errorf("compare() = %d, want 0", got)
	}
}

func TestConstraint_Compare_EqualMajorMinor(t *testing.T) {
	c := constraint{major: 1, minor: 2, minorSet: true}
	if got := c.compare(mustParseVersion(t, "1.2.9")); got != 0 {
		t.Errorf("compare() = %d, want 0", got)
	}
}

func TestConstraint_Compare_LessMajor(t *testing.T) {
	c := constraint{major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if got := c.compare(mustParseVersion(t, "2.0.0")); got != -1 {
		t.Errorf("compare() = %d, want -1", got)
	}
}

func TestConstraint_Compare_LessMinor(t *testing.T) {
	c := constraint{major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if got := c.compare(mustParseVersion(t, "1.1.0")); got != -1 {
		t.Errorf("compare() = %d, want -1", got)
	}
}

func TestConstraint_Compare_LessPatch(t *testing.T) {
	c := constraint{major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if got := c.compare(mustParseVersion(t, "1.0.1")); got != -1 {
		t.Errorf("compare() = %d, want -1", got)
	}
}

func TestConstraint_Compare_GreaterMajor(t *testing.T) {
	c := constraint{major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if got := c.compare(mustParseVersion(t, "1.0.0")); got != 1 {
		t.Errorf("compare() = %d, want 1", got)
	}
}

func TestConstraint_Compare_GreaterMinor(t *testing.T) {
	c := constraint{major: 1, minor: 5, patch: 0, minorSet: true, patchSet: true}
	if got := c.compare(mustParseVersion(t, "1.0.0")); got != 1 {
		t.Errorf("compare() = %d, want 1", got)
	}
}

func TestConstraint_Compare_GreaterPatch(t *testing.T) {
	c := constraint{major: 1, minor: 0, patch: 5, minorSet: true, patchSet: true}
	if got := c.compare(mustParseVersion(t, "1.0.0")); got != 1 {
		t.Errorf("compare() = %d, want 1", got)
	}
}

// Unknown operator test

func TestConstraint_Matches_UnknownOperator(t *testing.T) {
	c := constraint{operator: "??", major: 1}
	if c.matches(mustParseVersion(t, "1.0.0")) {
		t.Error("unknown operator should not match")
	}
}

// String tests

func TestConstraint_String_EqualMajorOnly(t *testing.T) {
	c := constraint{operator: "=", major: 1}
	if got := c.String(); got != "=1" {
		t.Errorf("String() = %q, want %q", got, "=1")
	}
}

func TestConstraint_String_EqualMajorMinor(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, minorSet: true}
	if got := c.String(); got != "=1.2" {
		t.Errorf("String() = %q, want %q", got, "=1.2")
	}
}

func TestConstraint_String_EqualFull(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if got := c.String(); got != "=1.2.3" {
		t.Errorf("String() = %q, want %q", got, "=1.2.3")
	}
}

func TestConstraint_String_GreaterThan(t *testing.T) {
	c := constraint{operator: ">", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if got := c.String(); got != ">1.0.0" {
		t.Errorf("String() = %q, want %q", got, ">1.0.0")
	}
}

func TestConstraint_String_GreaterThanOrEqual(t *testing.T) {
	c := constraint{operator: ">=", major: 2, minor: 1, patch: 0, minorSet: true, patchSet: true}
	if got := c.String(); got != ">=2.1.0" {
		t.Errorf("String() = %q, want %q", got, ">=2.1.0")
	}
}

func TestConstraint_String_LessThan(t *testing.T) {
	c := constraint{operator: "<", major: 3, minor: 0, patch: 0, minorSet: true, patchSet: true}
	if got := c.String(); got != "<3.0.0" {
		t.Errorf("String() = %q, want %q", got, "<3.0.0")
	}
}

func TestConstraint_String_LessThanOrEqual(t *testing.T) {
	c := constraint{operator: "<=", major: 2, minor: 5, patch: 0, minorSet: true, patchSet: true}
	if got := c.String(); got != "<=2.5.0" {
		t.Errorf("String() = %q, want %q", got, "<=2.5.0")
	}
}

func TestConstraint_String_NotEqual(t *testing.T) {
	c := constraint{operator: "!=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if got := c.String(); got != "!=1.2.3" {
		t.Errorf("String() = %q, want %q", got, "!=1.2.3")
	}
}

func TestConstraint_String_Tilde(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if got := c.String(); got != "~1.2.3" {
		t.Errorf("String() = %q, want %q", got, "~1.2.3")
	}
}

func TestConstraint_String_TildePartial(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, minorSet: true}
	if got := c.String(); got != "~1.2" {
		t.Errorf("String() = %q, want %q", got, "~1.2")
	}
}

func TestConstraint_String_Caret(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if got := c.String(); got != "^1.2.3" {
		t.Errorf("String() = %q, want %q", got, "^1.2.3")
	}
}

func TestConstraint_String_CaretZero(t *testing.T) {
	c := constraint{operator: "^", major: 0, minor: 2, patch: 3, minorSet: true, patchSet: true}
	if got := c.String(); got != "^0.2.3" {
		t.Errorf("String() = %q, want %q", got, "^0.2.3")
	}
}

func TestConstraint_String_CaretPartial(t *testing.T) {
	c := constraint{operator: "^", major: 1}
	if got := c.String(); got != "^1" {
		t.Errorf("String() = %q, want %q", got, "^1")
	}
}

// Prerelease versions never match constraints

func TestConstraint_Matches_PrereleaseNeverMatches(t *testing.T) {
	c := constraint{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true}
	v := mustParseVersion(t, "1.2.3-alpha.1")
	if c.matches(v) {
		t.Error("prerelease version should not match constraint")
	}
}

func TestConstraint_Matches_PrereleaseNeverMatchesRange(t *testing.T) {
	c := constraint{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	v := mustParseVersion(t, "1.5.0-beta.2")
	if c.matches(v) {
		t.Error("prerelease version should not match range constraint")
	}
}

func TestConstraint_Matches_PrereleaseNeverMatchesTilde(t *testing.T) {
	c := constraint{operator: "~", major: 1, minor: 2, patch: 0, minorSet: true, patchSet: true}
	v := mustParseVersion(t, "1.2.5-rc.1")
	if c.matches(v) {
		t.Error("prerelease version should not match tilde constraint")
	}
}

func TestConstraint_Matches_PrereleaseNeverMatchesCaret(t *testing.T) {
	c := constraint{operator: "^", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true}
	v := mustParseVersion(t, "1.9.0-alpha.1")
	if c.matches(v) {
		t.Error("prerelease version should not match caret constraint")
	}
}
