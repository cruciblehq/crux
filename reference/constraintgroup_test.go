package reference

import "testing"

// matches tests

func TestConstraintGroup_Matches_Empty(t *testing.T) {
	g := constraintGroup{}
	if !g.matches(mustParseVersion(t, "1.0.0")) {
		t.Error("empty group should match any version")
	}
}

func TestConstraintGroup_Matches_SingleConstraint(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true},
		},
	}
	if !g.matches(mustParseVersion(t, "1.2.3")) {
		t.Error("group with =1.2.3 should match 1.2.3")
	}
}

func TestConstraintGroup_Matches_SingleConstraint_Rejects(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true},
		},
	}
	if g.matches(mustParseVersion(t, "1.2.4")) {
		t.Error("group with =1.2.3 should not match 1.2.4")
	}
}

func TestConstraintGroup_Matches_AllMatch(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if !g.matches(mustParseVersion(t, "1.5.0")) {
		t.Error("group >=1.0.0 <2.0.0 should match 1.5.0")
	}
}

func TestConstraintGroup_Matches_FirstFails(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if g.matches(mustParseVersion(t, "0.9.0")) {
		t.Error("group >=1.0.0 <2.0.0 should not match 0.9.0")
	}
}

func TestConstraintGroup_Matches_SecondFails(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if g.matches(mustParseVersion(t, "2.0.0")) {
		t.Error("group >=1.0.0 <2.0.0 should not match 2.0.0")
	}
}

func TestConstraintGroup_Matches_ThreeConstraints_AllMatch(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "!=", major: 1, minor: 5, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if !g.matches(mustParseVersion(t, "1.4.0")) {
		t.Error("group >=1.0.0 <2.0.0 !=1.5.0 should match 1.4.0")
	}
}

func TestConstraintGroup_Matches_ThreeConstraints_ThirdFails(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "!=", major: 1, minor: 5, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if g.matches(mustParseVersion(t, "1.5.0")) {
		t.Error("group >=1.0.0 <2.0.0 !=1.5.0 should not match 1.5.0")
	}
}

func TestConstraintGroup_Matches_LowerBoundary(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if !g.matches(mustParseVersion(t, "1.0.0")) {
		t.Error("group >=1.0.0 <2.0.0 should match 1.0.0 (lower boundary)")
	}
}

func TestConstraintGroup_Matches_UpperBoundary(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if !g.matches(mustParseVersion(t, "1.9.9")) {
		t.Error("group >=1.0.0 <2.0.0 should match 1.9.9 (just below upper boundary)")
	}
}

// String tests

func TestConstraintGroup_String_SingleConstraint(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: "=", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true},
		},
	}
	if got := g.String(); got != "=1.2.3" {
		t.Errorf("String() = %q, want %q", got, "=1.2.3")
	}
}

func TestConstraintGroup_String_TwoConstraints(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if got := g.String(); got != ">=1.0.0 <2.0.0" {
		t.Errorf("String() = %q, want %q", got, ">=1.0.0 <2.0.0")
	}
}

func TestConstraintGroup_String_ThreeConstraints(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "<", major: 2, minor: 0, patch: 0, minorSet: true, patchSet: true},
			{operator: "!=", major: 1, minor: 5, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if got := g.String(); got != ">=1.0.0 <2.0.0 !=1.5.0" {
		t.Errorf("String() = %q, want %q", got, ">=1.0.0 <2.0.0 !=1.5.0")
	}
}

func TestConstraintGroup_String_TildeConstraint(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: "~", major: 1, minor: 2, patch: 0, minorSet: true, patchSet: true},
		},
	}
	if got := g.String(); got != "~1.2.0" {
		t.Errorf("String() = %q, want %q", got, "~1.2.0")
	}
}

func TestConstraintGroup_String_CaretConstraint(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: "^", major: 1, minor: 2, patch: 3, minorSet: true, patchSet: true},
		},
	}
	if got := g.String(); got != "^1.2.3" {
		t.Errorf("String() = %q, want %q", got, "^1.2.3")
	}
}

func TestConstraintGroup_String_PartialVersion(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1, minor: 2, minorSet: true},
		},
	}
	if got := g.String(); got != ">=1.2" {
		t.Errorf("String() = %q, want %q", got, ">=1.2")
	}
}

func TestConstraintGroup_String_MajorOnly(t *testing.T) {
	g := constraintGroup{
		constraints: []constraint{
			{operator: ">=", major: 1},
		},
	}
	if got := g.String(); got != ">=1" {
		t.Errorf("String() = %q, want %q", got, ">=1")
	}
}
