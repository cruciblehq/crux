package reference

import "testing"

func mustParseConstraint(t *testing.T, s string) *VersionConstraint {
	t.Helper()
	vc, err := ParseVersionConstraint(s)
	if err != nil {
		t.Fatalf("unexpected error parsing %q: %v", s, err)
	}
	return vc
}

func mustMatch(t *testing.T, vc *VersionConstraint, version string) bool {
	t.Helper()
	match, err := vc.Matches(version)
	if err != nil {
		t.Fatalf("unexpected error matching %q: %v", version, err)
	}
	return match
}

// ParseVersionConstraint tests

func TestParseVersionConstraint_ExactVersion(t *testing.T) {
	vc := mustParseConstraint(t, "1.2.3")
	if vc.String() != "=1.2.3" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.2.3")
	}
}

func TestParseVersionConstraint_ExactVersionWithEquals(t *testing.T) {
	vc := mustParseConstraint(t, "=1.2.3")
	if vc.String() != "=1.2.3" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.2.3")
	}
}

func TestParseVersionConstraint_PartialVersion(t *testing.T) {
	vc := mustParseConstraint(t, "1.2")
	if vc.String() != "=1.2" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.2")
	}
}

func TestParseVersionConstraint_MajorOnly(t *testing.T) {
	vc := mustParseConstraint(t, "1")
	if vc.String() != "=1" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1")
	}
}

func TestParseVersionConstraint_Empty(t *testing.T) {
	_, err := ParseVersionConstraint("")
	if err == nil {
		t.Error("expected error for empty constraint")
	}
}

func TestParseVersionConstraint_Whitespace(t *testing.T) {
	_, err := ParseVersionConstraint("   ")
	if err == nil {
		t.Error("expected error for whitespace-only constraint")
	}
}

// Comparison operator tests

func TestParseVersionConstraint_GreaterThanWithUpperBound(t *testing.T) {
	vc := mustParseConstraint(t, ">1.0.0 <2.0.0")
	if vc.String() != ">1.0.0 <2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">1.0.0 <2.0.0")
	}
}

func TestParseVersionConstraint_GreaterThanOrEqualWithUpperBound(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0")
	if vc.String() != ">=1.0.0 <2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <2.0.0")
	}
}

func TestParseVersionConstraint_LessThan(t *testing.T) {
	vc := mustParseConstraint(t, "<2.0.0")
	if vc.String() != "<2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), "<2.0.0")
	}
}

func TestParseVersionConstraint_LessThanOrEqual(t *testing.T) {
	vc := mustParseConstraint(t, "<=2.0.0")
	if vc.String() != "<=2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), "<=2.0.0")
	}
}

func TestParseVersionConstraint_NotEqual(t *testing.T) {
	vc := mustParseConstraint(t, "!=1.5.0")
	if vc.String() != "!=1.5.0" {
		t.Errorf("String() = %q, want %q", vc.String(), "!=1.5.0")
	}
}

// Unbounded range rejection tests

func TestParseVersionConstraint_UnboundedGreaterThan(t *testing.T) {
	_, err := ParseVersionConstraint(">1.0.0")
	if err == nil {
		t.Error("expected error for unbounded >1.0.0")
	}
}

func TestParseVersionConstraint_UnboundedGreaterThanOrEqual(t *testing.T) {
	_, err := ParseVersionConstraint(">=1.0.0")
	if err == nil {
		t.Error("expected error for unbounded >=1.0.0")
	}
}

// Tilde and caret tests

func TestParseVersionConstraint_Tilde(t *testing.T) {
	vc := mustParseConstraint(t, "~1.2.3")
	if vc.String() != "~1.2.3" {
		t.Errorf("String() = %q, want %q", vc.String(), "~1.2.3")
	}
}

func TestParseVersionConstraint_Caret(t *testing.T) {
	vc := mustParseConstraint(t, "^1.2.3")
	if vc.String() != "^1.2.3" {
		t.Errorf("String() = %q, want %q", vc.String(), "^1.2.3")
	}
}

func TestParseVersionConstraint_CaretZeroMajor(t *testing.T) {
	vc := mustParseConstraint(t, "^0.2.3")

	if !mustMatch(t, vc, "0.2.3") {
		t.Error("^0.2.3 should match 0.2.3")
	}
	if !mustMatch(t, vc, "0.2.9") {
		t.Error("^0.2.3 should match 0.2.9")
	}
	if mustMatch(t, vc, "0.3.0") {
		t.Error("^0.2.3 should not match 0.3.0")
	}
}

// Hyphen range tests

func TestParseVersionConstraint_HyphenRange(t *testing.T) {
	vc := mustParseConstraint(t, "1.0.0 - 2.0.0")
	if vc.String() != ">=1.0.0 <=2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <=2.0.0")
	}
}

func TestParseVersionConstraint_HyphenRangeWithConstraintBefore(t *testing.T) {
	vc := mustParseConstraint(t, "!=1.5.0 1.0.0 - 2.0.0")
	if vc.String() != "!=1.5.0 >=1.0.0 <=2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), "!=1.5.0 >=1.0.0 <=2.0.0")
	}
}

func TestParseVersionConstraint_HyphenRangeWithConstraintAfter(t *testing.T) {
	vc := mustParseConstraint(t, "1.0.0 - 2.0.0 !=1.5.0")
	if vc.String() != ">=1.0.0 <=2.0.0 !=1.5.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <=2.0.0 !=1.5.0")
	}
}

func TestParseVersionConstraint_MultipleHyphenRanges(t *testing.T) {
	vc := mustParseConstraint(t, "1.0.0 - 2.0.0 3.0.0 - 4.0.0")
	if vc.String() != ">=1.0.0 <=2.0.0 >=3.0.0 <=4.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <=2.0.0 >=3.0.0 <=4.0.0")
	}
}

func TestParseVersionConstraint_HyphenRangeMissingLowerBound(t *testing.T) {
	_, err := ParseVersionConstraint("- 2.0.0")
	if err == nil {
		t.Error("expected error for hyphen range missing lower bound")
	}
}

func TestParseVersionConstraint_HyphenRangeMissingUpperBound(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0 -")
	if err == nil {
		t.Error("expected error for hyphen range missing upper bound")
	}
}

func TestParseVersionConstraint_HyphenRangeConsecutiveHyphens(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0 - - 2.0.0")
	if err == nil {
		t.Error("expected error for consecutive hyphens")
	}
}

func TestParseVersionConstraint_HyphenRangeLowerBoundWithOperator(t *testing.T) {
	_, err := ParseVersionConstraint(">=1.0.0 - 2.0.0")
	if err == nil {
		t.Error("expected error for lower bound with operator")
	}
}

func TestParseVersionConstraint_HyphenRangeUpperBoundWithOperator(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0 - <=2.0.0")
	if err == nil {
		t.Error("expected error for upper bound with operator")
	}
}

func TestParseVersionConstraint_HyphenRangeWithWildcardLower(t *testing.T) {
	_, err := ParseVersionConstraint("1.x - 2.0.0")
	if err == nil {
		t.Error("expected error for wildcard in lower bound")
	}
}

func TestParseVersionConstraint_HyphenRangeWithWildcardUpper(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0 - 2.x")
	if err == nil {
		t.Error("expected error for wildcard in upper bound")
	}
}

// Wildcard tests

func TestParseVersionConstraint_WildcardMajorX(t *testing.T) {
	vc := mustParseConstraint(t, "1.x")
	if vc.String() != "=1" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1")
	}
}

func TestParseVersionConstraint_WildcardMinorX(t *testing.T) {
	vc := mustParseConstraint(t, "1.2.x")
	if vc.String() != "=1.2" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.2")
	}
}

func TestParseVersionConstraint_WildcardUppercaseX(t *testing.T) {
	vc := mustParseConstraint(t, "1.X")
	if vc.String() != "=1" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1")
	}
}

func TestParseVersionConstraint_WildcardStar(t *testing.T) {
	_, err := ParseVersionConstraint("*")
	if err == nil {
		t.Error("expected error for bare wildcard *")
	}
}

func TestParseVersionConstraint_WildcardBareX(t *testing.T) {
	_, err := ParseVersionConstraint("x")
	if err == nil {
		t.Error("expected error for bare wildcard x")
	}
}

func TestParseVersionConstraint_WildcardMultiple(t *testing.T) {
	_, err := ParseVersionConstraint("1.x.x")
	if err == nil {
		t.Error("expected error for multiple wildcards")
	}
}

func TestParseVersionConstraint_WildcardWithOperator(t *testing.T) {
	_, err := ParseVersionConstraint(">1.x")
	if err == nil {
		t.Error("expected error for wildcard with operator")
	}
}

// OR operator tests

func TestParseVersionConstraint_OrOperator(t *testing.T) {
	vc := mustParseConstraint(t, "1.2.3 || 2.0.0")
	if vc.String() != "=1.2.3 || =2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.2.3 || =2.0.0")
	}
}

func TestParseVersionConstraint_OrOperatorMultiple(t *testing.T) {
	vc := mustParseConstraint(t, "=1.0.0 || =2.0.0 || =3.0.0")
	if vc.String() != "=1.0.0 || =2.0.0 || =3.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.0.0 || =2.0.0 || =3.0.0")
	}
}

func TestParseVersionConstraint_OrOperatorWithRanges(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0 || >=3.0.0 <4.0.0")
	if vc.String() != ">=1.0.0 <2.0.0 || >=3.0.0 <4.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <2.0.0 || >=3.0.0 <4.0.0")
	}
}

func TestParseVersionConstraint_OrOperatorEmptyLeft(t *testing.T) {
	_, err := ParseVersionConstraint("|| 1.0.0")
	if err == nil {
		t.Error("expected error for empty left side of OR")
	}
}

func TestParseVersionConstraint_OrOperatorEmptyRight(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0 ||")
	if err == nil {
		t.Error("expected error for empty right side of OR")
	}
}

func TestParseVersionConstraint_OrOperatorEmptyMiddle(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0 || || 2.0.0")
	if err == nil {
		t.Error("expected error for empty middle in OR expression")
	}
}

// AND operator tests

func TestParseVersionConstraint_AndOperator(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0 !=1.5.0")
	if vc.String() != ">=1.0.0 <2.0.0 !=1.5.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <2.0.0 !=1.5.0")
	}
}

// String method tests

func TestVersionConstraint_String_Nil(t *testing.T) {
	var vc *VersionConstraint
	if vc.String() != "" {
		t.Errorf("String() = %q, want %q", vc.String(), "")
	}
}

func TestVersionConstraint_String_Empty(t *testing.T) {
	vc := &VersionConstraint{}
	if vc.String() != "" {
		t.Errorf("String() = %q, want %q", vc.String(), "")
	}
}

// Matches method tests

func TestVersionConstraint_Matches_Exact(t *testing.T) {
	vc := mustParseConstraint(t, "1.2.3")
	if !mustMatch(t, vc, "1.2.3") {
		t.Error("1.2.3 should match 1.2.3")
	}
}

func TestVersionConstraint_Matches_ExactNoMatch(t *testing.T) {
	vc := mustParseConstraint(t, "1.2.3")
	if mustMatch(t, vc, "1.2.4") {
		t.Error("1.2.3 should not match 1.2.4")
	}
}

func TestVersionConstraint_Matches_Range(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0")
	if !mustMatch(t, vc, "1.5.0") {
		t.Error(">=1.0.0 <2.0.0 should match 1.5.0")
	}
}

func TestVersionConstraint_Matches_RangeLowerBound(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0")
	if !mustMatch(t, vc, "1.0.0") {
		t.Error(">=1.0.0 <2.0.0 should match 1.0.0")
	}
}

func TestVersionConstraint_Matches_RangeUpperBound(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0")
	if mustMatch(t, vc, "2.0.0") {
		t.Error(">=1.0.0 <2.0.0 should not match 2.0.0")
	}
}

func TestVersionConstraint_Matches_RangeBelowLower(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0")
	if mustMatch(t, vc, "0.9.0") {
		t.Error(">=1.0.0 <2.0.0 should not match 0.9.0")
	}
}

func TestVersionConstraint_Matches_HyphenRange(t *testing.T) {
	vc := mustParseConstraint(t, "1.0.0 - 2.0.0")
	if !mustMatch(t, vc, "1.5.0") {
		t.Error("1.0.0 - 2.0.0 should match 1.5.0")
	}
}

func TestVersionConstraint_Matches_HyphenRangeUpperInclusive(t *testing.T) {
	vc := mustParseConstraint(t, "1.0.0 - 2.0.0")
	if !mustMatch(t, vc, "2.0.0") {
		t.Error("1.0.0 - 2.0.0 should match 2.0.0 (inclusive)")
	}
}

func TestVersionConstraint_Matches_Or(t *testing.T) {
	vc := mustParseConstraint(t, "1.0.0 || 2.0.0")

	if !mustMatch(t, vc, "1.0.0") {
		t.Error("1.0.0 || 2.0.0 should match 1.0.0")
	}
	if !mustMatch(t, vc, "2.0.0") {
		t.Error("1.0.0 || 2.0.0 should match 2.0.0")
	}
	if mustMatch(t, vc, "1.5.0") {
		t.Error("1.0.0 || 2.0.0 should not match 1.5.0")
	}
}

func TestVersionConstraint_Matches_Tilde(t *testing.T) {
	vc := mustParseConstraint(t, "~1.2.3")

	if !mustMatch(t, vc, "1.2.3") {
		t.Error("~1.2.3 should match 1.2.3")
	}
	if !mustMatch(t, vc, "1.2.9") {
		t.Error("~1.2.3 should match 1.2.9")
	}
	if mustMatch(t, vc, "1.3.0") {
		t.Error("~1.2.3 should not match 1.3.0")
	}
}

func TestVersionConstraint_Matches_Caret(t *testing.T) {
	vc := mustParseConstraint(t, "^1.2.3")

	if !mustMatch(t, vc, "1.2.3") {
		t.Error("^1.2.3 should match 1.2.3")
	}
	if !mustMatch(t, vc, "1.9.0") {
		t.Error("^1.2.3 should match 1.9.0")
	}
	if mustMatch(t, vc, "2.0.0") {
		t.Error("^1.2.3 should not match 2.0.0")
	}
}

func TestVersionConstraint_Matches_NotEqual(t *testing.T) {
	vc := mustParseConstraint(t, "!=1.5.0")

	if mustMatch(t, vc, "1.5.0") {
		t.Error("!=1.5.0 should not match 1.5.0")
	}
	if !mustMatch(t, vc, "1.5.1") {
		t.Error("!=1.5.0 should match 1.5.1")
	}
}

func TestParseVersionConstraint_Complex(t *testing.T) {
	vc := mustParseConstraint(t, ">=1.0.0 <2.0.0 !=1.5.0 || ^3.0.0")

	if !mustMatch(t, vc, "1.2.0") {
		t.Error("should match 1.2.0")
	}
	if mustMatch(t, vc, "1.5.0") {
		t.Error("should not match 1.5.0 (excluded)")
	}
	if !mustMatch(t, vc, "3.5.0") {
		t.Error("should match 3.5.0")
	}
	if mustMatch(t, vc, "4.0.0") {
		t.Error("should not match 4.0.0")
	}
}

func TestParseVersionConstraint_HyphenRangeWithExclusion(t *testing.T) {
	vc := mustParseConstraint(t, "1.0.0 - 2.0.0 !=1.5.0")

	if !mustMatch(t, vc, "1.2.0") {
		t.Error("should match 1.2.0")
	}
	if mustMatch(t, vc, "1.5.0") {
		t.Error("should not match 1.5.0 (excluded)")
	}
	if !mustMatch(t, vc, "2.0.0") {
		t.Error("should match 2.0.0")
	}
}

// Whitespace normalization tests

func TestParseVersionConstraint_ExtraWhitespace(t *testing.T) {
	vc := mustParseConstraint(t, "  >=1.0.0   <2.0.0  ")
	if vc.String() != ">=1.0.0 <2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <2.0.0")
	}
}

func TestParseVersionConstraint_OrWithExtraWhitespace(t *testing.T) {
	vc := mustParseConstraint(t, "  1.0.0  ||  2.0.0  ")
	if vc.String() != "=1.0.0 || =2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.0.0 || =2.0.0")
	}
}

func TestParseVersionConstraint_VPrefix(t *testing.T) {
	vc := mustParseConstraint(t, "v1.2.3")
	if vc.String() != "=1.2.3" {
		t.Errorf("String() = %q, want %q", vc.String(), "=1.2.3")
	}
}

func TestParseVersionConstraint_VPrefixWithOperator(t *testing.T) {
	vc := mustParseConstraint(t, ">=v1.0.0 <v2.0.0")
	if vc.String() != ">=1.0.0 <2.0.0" {
		t.Errorf("String() = %q, want %q", vc.String(), ">=1.0.0 <2.0.0")
	}
}

// Replace the prerelease tests with rejection tests:

func TestParseVersionConstraint_PrereleaseRejected(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0-alpha")
	if err == nil {
		t.Error("expected error for prerelease version")
	}
}

func TestParseVersionConstraint_PrereleaseWithOperatorRejected(t *testing.T) {
	_, err := ParseVersionConstraint(">=1.0.0-alpha <2.0.0")
	if err == nil {
		t.Error("expected error for prerelease version with operator")
	}
}

func TestParseVersionConstraint_PrereleaseInHyphenRangeRejected(t *testing.T) {
	_, err := ParseVersionConstraint("1.0.0-alpha - 2.0.0")
	if err == nil {
		t.Error("expected error for prerelease in hyphen range")
	}
}
