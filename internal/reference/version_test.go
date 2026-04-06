package reference

import "testing"

func mustParseVersion(t *testing.T, s string) *Version {
	t.Helper()
	v, err := ParseVersion(s)
	if err != nil {
		t.Fatalf("ParseVersion(%q) failed: %v", s, err)
	}
	return v
}

// ParseVersion tests

func TestParseVersion_Basic(t *testing.T) {
	v := mustParseVersion(t, "1.2.3")
	if v.Major != 1 {
		t.Errorf("Major = %d, want 1", v.Major)
	}
	if v.Minor != 2 {
		t.Errorf("Minor = %d, want 2", v.Minor)
	}
	if v.Patch != 3 {
		t.Errorf("Patch = %d, want 3", v.Patch)
	}
}

func TestParseVersion_WithVPrefix(t *testing.T) {
	v := mustParseVersion(t, "v1.2.3")
	if v.Major != 1 {
		t.Errorf("Major = %d, want 1", v.Major)
	}
}

func TestParseVersion_WithUppercaseVPrefix(t *testing.T) {
	v := mustParseVersion(t, "V1.2.3")
	if v.Major != 1 {
		t.Errorf("Major = %d, want 1", v.Major)
	}
}

func TestParseVersion_WithPrerelease(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-alpha.1")
	if v.Prerelease != "alpha.1" {
		t.Errorf("Prerelease = %q, want %q", v.Prerelease, "alpha.1")
	}
}

func TestParseVersion_WithPrereleaseZero(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-alpha.0")
	if v.Prerelease != "alpha.0" {
		t.Errorf("Prerelease = %q, want %q", v.Prerelease, "alpha.0")
	}
}

func TestParseVersion_WithPrereleaseBeta(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-beta.2")
	if v.Prerelease != "beta.2" {
		t.Errorf("Prerelease = %q, want %q", v.Prerelease, "beta.2")
	}
}

func TestParseVersion_WithPrereleaseRc(t *testing.T) {
	v := mustParseVersion(t, "1.0.0-rc.1")
	if v.Prerelease != "rc.1" {
		t.Errorf("Prerelease = %q, want %q", v.Prerelease, "rc.1")
	}
}

func TestParseVersion_WithPrereleaseLargeNumber(t *testing.T) {
	v := mustParseVersion(t, "1.0.0-alpha.123")
	if v.Prerelease != "alpha.123" {
		t.Errorf("Prerelease = %q, want %q", v.Prerelease, "alpha.123")
	}
}

func TestParseVersion_WithBuild(t *testing.T) {
	v := mustParseVersion(t, "1.2.3+build.456")
	if v.Build != "build.456" {
		t.Errorf("Build = %q, want %q", v.Build, "build.456")
	}
}

func TestParseVersion_WithBuildNumeric(t *testing.T) {
	v := mustParseVersion(t, "1.2.3+001")
	if v.Build != "001" {
		t.Errorf("Build = %q, want %q", v.Build, "001")
	}
}

func TestParseVersion_WithPrereleaseAndBuild(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-alpha.1+001")
	if v.Prerelease != "alpha.1" {
		t.Errorf("Prerelease = %q, want %q", v.Prerelease, "alpha.1")
	}
	if v.Build != "001" {
		t.Errorf("Build = %q, want %q", v.Build, "001")
	}
}

func TestParseVersion_ZeroVersion(t *testing.T) {
	v := mustParseVersion(t, "0.0.0")
	if v.Major != 0 || v.Minor != 0 || v.Patch != 0 {
		t.Errorf("got %d.%d.%d, want 0.0.0", v.Major, v.Minor, v.Patch)
	}
}

func TestParseVersion_LargeNumbers(t *testing.T) {
	v := mustParseVersion(t, "100.200.300")
	if v.Major != 100 {
		t.Errorf("Major = %d, want 100", v.Major)
	}
	if v.Minor != 200 {
		t.Errorf("Minor = %d, want 200", v.Minor)
	}
	if v.Patch != 300 {
		t.Errorf("Patch = %d, want 300", v.Patch)
	}
}

func TestParseVersion_TrimsWhitespace(t *testing.T) {
	v := mustParseVersion(t, "  1.2.3  ")
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("got %d.%d.%d, want 1.2.3", v.Major, v.Minor, v.Patch)
	}
}

// ParseVersion error tests

func TestParseVersion_Empty(t *testing.T) {
	_, err := ParseVersion("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestParseVersion_Whitespace(t *testing.T) {
	_, err := ParseVersion("   ")
	if err == nil {
		t.Error("expected error for whitespace-only string")
	}
}

func TestParseVersion_MajorOnly(t *testing.T) {
	_, err := ParseVersion("1")
	if err == nil {
		t.Error("expected error for major-only version")
	}
}

func TestParseVersion_MajorMinorOnly(t *testing.T) {
	_, err := ParseVersion("1.2")
	if err == nil {
		t.Error("expected error for major.minor-only version")
	}
}

func TestParseVersion_TooManyParts(t *testing.T) {
	_, err := ParseVersion("1.2.3.4")
	if err == nil {
		t.Error("expected error for too many parts")
	}
}

func TestParseVersion_InvalidMajor(t *testing.T) {
	_, err := ParseVersion("a.2.3")
	if err == nil {
		t.Error("expected error for invalid major")
	}
}

func TestParseVersion_InvalidMinor(t *testing.T) {
	_, err := ParseVersion("1.b.3")
	if err == nil {
		t.Error("expected error for invalid minor")
	}
}

func TestParseVersion_InvalidPatch(t *testing.T) {
	_, err := ParseVersion("1.2.c")
	if err == nil {
		t.Error("expected error for invalid patch")
	}
}

func TestParseVersion_NegativeMajor(t *testing.T) {
	_, err := ParseVersion("-1.2.3")
	if err == nil {
		t.Error("expected error for negative major")
	}
}

func TestParseVersion_NegativeMinor(t *testing.T) {
	_, err := ParseVersion("1.-2.3")
	if err == nil {
		t.Error("expected error for negative minor")
	}
}

func TestParseVersion_NegativePatch(t *testing.T) {
	_, err := ParseVersion("1.2.-3")
	if err == nil {
		t.Error("expected error for negative patch")
	}
}

func TestParseVersion_LeadingDot(t *testing.T) {
	_, err := ParseVersion(".2.3")
	if err == nil {
		t.Error("expected error for leading dot")
	}
}

func TestParseVersion_TrailingDot(t *testing.T) {
	_, err := ParseVersion("1.2.")
	if err == nil {
		t.Error("expected error for trailing dot")
	}
}

func TestParseVersion_DoubleDot(t *testing.T) {
	_, err := ParseVersion("1..3")
	if err == nil {
		t.Error("expected error for double dot")
	}
}

func TestParseVersion_PrereleaseWithoutNumber(t *testing.T) {
	_, err := ParseVersion("1.2.3-alpha")
	if err == nil {
		t.Error("expected error for prerelease without number")
	}
}

func TestParseVersion_PrereleaseStartsWithNumber(t *testing.T) {
	_, err := ParseVersion("1.2.3-1.2")
	if err == nil {
		t.Error("expected error for prerelease starting with number")
	}
}

func TestParseVersion_PrereleaseLeadingZero(t *testing.T) {
	_, err := ParseVersion("1.2.3-alpha.01")
	if err == nil {
		t.Error("expected error for prerelease with leading zero")
	}
}

func TestParseVersion_PrereleaseMultipleLeadingZeros(t *testing.T) {
	_, err := ParseVersion("1.2.3-alpha.007")
	if err == nil {
		t.Error("expected error for prerelease with multiple leading zeros")
	}
}

func TestParseVersion_PrereleaseTooManySegments(t *testing.T) {
	_, err := ParseVersion("1.2.3-alpha.1.2")
	if err == nil {
		t.Error("expected error for prerelease with too many segments")
	}
}

func TestParseVersion_PrereleaseWithHyphen(t *testing.T) {
	_, err := ParseVersion("1.2.3-alpha-beta.1")
	if err == nil {
		t.Error("expected error for prerelease with hyphen")
	}
}

func TestParseVersion_EmptyPrerelease(t *testing.T) {
	_, err := ParseVersion("1.2.3-")
	if err == nil {
		t.Error("expected error for empty prerelease")
	}
}

func TestParseVersion_EmptyBuild(t *testing.T) {
	_, err := ParseVersion("1.2.3+")
	if err == nil {
		t.Error("expected error for empty build")
	}
}

func TestParseVersion_InvalidBuildCharacter(t *testing.T) {
	_, err := ParseVersion("1.2.3+build#meta")
	if err == nil {
		t.Error("expected error for invalid build character")
	}
}

// Additional edge case tests

func TestParseVersion_LeadingZeroInMajor(t *testing.T) {
	// Leading zeros in numeric identifiers are valid in Go's Atoi
	// but semver spec says they should not have leading zeros
	v := mustParseVersion(t, "01.2.3")
	if v.Major != 1 {
		t.Errorf("Major = %d, want 1", v.Major)
	}
}

func TestParseVersion_LeadingZeroInMinor(t *testing.T) {
	v := mustParseVersion(t, "1.02.3")
	if v.Minor != 2 {
		t.Errorf("Minor = %d, want 2", v.Minor)
	}
}

func TestParseVersion_LeadingZeroInPatch(t *testing.T) {
	v := mustParseVersion(t, "1.2.03")
	if v.Patch != 3 {
		t.Errorf("Patch = %d, want 3", v.Patch)
	}
}

func TestParseVersion_BuildWithHyphen(t *testing.T) {
	v := mustParseVersion(t, "1.2.3+build-info")
	if v.Build != "build-info" {
		t.Errorf("Build = %q, want %q", v.Build, "build-info")
	}
}

func TestParseVersion_BuildWithMultipleDots(t *testing.T) {
	v := mustParseVersion(t, "1.2.3+build.info.123")
	if v.Build != "build.info.123" {
		t.Errorf("Build = %q, want %q", v.Build, "build.info.123")
	}
}

func TestParseVersion_PrereleaseAndBuildWithHyphen(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-alpha.1+build-info")
	if v.Prerelease != "alpha.1" {
		t.Errorf("Prerelease = %q, want %q", v.Prerelease, "alpha.1")
	}
	if v.Build != "build-info" {
		t.Errorf("Build = %q, want %q", v.Build, "build-info")
	}
}

// String tests

func TestVersion_String_Basic(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3}
	if v.String() != "1.2.3" {
		t.Errorf("String() = %q, want %q", v.String(), "1.2.3")
	}
}

func TestVersion_String_WithPrerelease(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"}
	if v.String() != "1.2.3-alpha.1" {
		t.Errorf("String() = %q, want %q", v.String(), "1.2.3-alpha.1")
	}
}

func TestVersion_String_WithBuild(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3, Build: "001"}
	if v.String() != "1.2.3+001" {
		t.Errorf("String() = %q, want %q", v.String(), "1.2.3+001")
	}
}

func TestVersion_String_WithPrereleaseAndBuild(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1", Build: "001"}
	if v.String() != "1.2.3-alpha.1+001" {
		t.Errorf("String() = %q, want %q", v.String(), "1.2.3-alpha.1+001")
	}
}

func TestVersion_String_Zero(t *testing.T) {
	v := &Version{Major: 0, Minor: 0, Patch: 0}
	if v.String() != "0.0.0" {
		t.Errorf("String() = %q, want %q", v.String(), "0.0.0")
	}
}

// IsPrerelease tests

func TestVersion_IsPrerelease_Stable(t *testing.T) {
	v := mustParseVersion(t, "1.2.3")
	if v.IsPrerelease() {
		t.Error("expected stable version to not be prerelease")
	}
}

func TestVersion_IsPrerelease_Alpha(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-alpha.1")
	if !v.IsPrerelease() {
		t.Error("expected alpha version to be prerelease")
	}
}

func TestVersion_IsPrerelease_WithBuildOnly(t *testing.T) {
	v := mustParseVersion(t, "1.2.3+build")
	if v.IsPrerelease() {
		t.Error("expected version with only build metadata to not be prerelease")
	}
}

// Compare tests

func TestVersion_Compare_MajorLess(t *testing.T) {
	a := mustParseVersion(t, "1.0.0")
	b := mustParseVersion(t, "2.0.0")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_MajorGreater(t *testing.T) {
	a := mustParseVersion(t, "2.0.0")
	b := mustParseVersion(t, "1.0.0")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 1 {
		t.Errorf("Compare() = %d, want 1", cmp)
	}
}

func TestVersion_Compare_MinorLess(t *testing.T) {
	a := mustParseVersion(t, "1.1.0")
	b := mustParseVersion(t, "1.2.0")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_MinorGreater(t *testing.T) {
	a := mustParseVersion(t, "1.2.0")
	b := mustParseVersion(t, "1.1.0")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 1 {
		t.Errorf("Compare() = %d, want 1", cmp)
	}
}

func TestVersion_Compare_PatchLess(t *testing.T) {
	a := mustParseVersion(t, "1.0.1")
	b := mustParseVersion(t, "1.0.2")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_PatchGreater(t *testing.T) {
	a := mustParseVersion(t, "1.0.2")
	b := mustParseVersion(t, "1.0.1")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 1 {
		t.Errorf("Compare() = %d, want 1", cmp)
	}
}

func TestVersion_Compare_Equal(t *testing.T) {
	a := mustParseVersion(t, "1.2.3")
	b := mustParseVersion(t, "1.2.3")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 0 {
		t.Errorf("Compare() = %d, want 0", cmp)
	}
}

func TestVersion_Compare_StableGreaterThanPrerelease(t *testing.T) {
	a := mustParseVersion(t, "1.0.0")
	b := mustParseVersion(t, "1.0.0-alpha.1")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 1 {
		t.Errorf("Compare() = %d, want 1", cmp)
	}
}

func TestVersion_Compare_PrereleaseLessThanStable(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1")
	b := mustParseVersion(t, "1.0.0")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_SamePrereleaseNumberLess(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1")
	b := mustParseVersion(t, "1.0.0-alpha.2")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_SamePrereleaseNumberGreater(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.2")
	b := mustParseVersion(t, "1.0.0-alpha.1")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 1 {
		t.Errorf("Compare() = %d, want 1", cmp)
	}
}

func TestVersion_Compare_SamePrereleaseEqual(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1")
	b := mustParseVersion(t, "1.0.0-alpha.1")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 0 {
		t.Errorf("Compare() = %d, want 0", cmp)
	}
}

func TestVersion_Compare_SamePrereleaseZero(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.0")
	b := mustParseVersion(t, "1.0.0-alpha.1")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_DifferentPrereleaseNotComparable(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1")
	b := mustParseVersion(t, "1.0.0-beta.1")
	_, valid := a.Compare(b)
	if valid {
		t.Error("expected invalid comparison for different prerelease identifiers")
	}
}

func TestVersion_Compare_DifferentPrereleaseAlphaRc(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1")
	b := mustParseVersion(t, "1.0.0-rc.1")
	_, valid := a.Compare(b)
	if valid {
		t.Error("expected invalid comparison for alpha vs rc")
	}
}

func TestVersion_Compare_DifferentMajorWithPrereleases(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1")
	b := mustParseVersion(t, "2.0.0-alpha.1")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_DifferentMinorWithPrereleases(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1")
	b := mustParseVersion(t, "1.1.0-alpha.1")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}
}

func TestVersion_Compare_BuildMetadataIgnored(t *testing.T) {
	a := mustParseVersion(t, "1.0.0+build1")
	b := mustParseVersion(t, "1.0.0+build2")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 0 {
		t.Errorf("Compare() = %d, want 0", cmp)
	}
}

func TestVersion_Compare_PrereleaseWithBuildMetadataIgnored(t *testing.T) {
	a := mustParseVersion(t, "1.0.0-alpha.1+build1")
	b := mustParseVersion(t, "1.0.0-alpha.1+build2")
	cmp, valid := a.Compare(b)
	if !valid {
		t.Error("expected valid comparison")
	}
	if cmp != 0 {
		t.Errorf("Compare() = %d, want 0", cmp)
	}
}

// IsPrerelease additional tests

func TestVersion_IsPrerelease_Beta(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-beta.1")
	if !v.IsPrerelease() {
		t.Error("expected beta version to be prerelease")
	}
}

func TestVersion_IsPrerelease_Rc(t *testing.T) {
	v := mustParseVersion(t, "1.2.3-rc.1")
	if !v.IsPrerelease() {
		t.Error("expected rc version to be prerelease")
	}
}
