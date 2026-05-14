package manifest

import "testing"

func TestIsValidName(t *testing.T) {
	cases := []string{"a", "foo", "bar_baz", "x1", "z99_abc", "abc_123_def"}
	for _, s := range cases {
		if !isValidName(s) {
			t.Errorf("isValidName(%q) = false, want true", s)
		}
	}
}

func TestIsValidNameInvalid(t *testing.T) {
	cases := []string{"", "A", "Foo", "1foo", "foo-bar", "foo bar", "_foo", "FOO", "foo!", "foo.bar"}
	for _, s := range cases {
		if isValidName(s) {
			t.Errorf("isValidName(%q) = true, want false", s)
		}
	}
}

func TestIsValidPlatform(t *testing.T) {
	cases := []string{"linux/amd64", "linux/arm64", "linux/arm64/v8", "darwin/amd64", "windows/amd64"}
	for _, s := range cases {
		if !isValidPlatform(s) {
			t.Errorf("isValidPlatform(%q) = false, want true", s)
		}
	}
}

func TestIsValidPlatformInvalid(t *testing.T) {
	cases := []string{"", "linux", "linux/", "/amd64", "Linux/amd64", "linux/AMD64", "linux-amd64", "linux/arm64-v8"}
	for _, s := range cases {
		if isValidPlatform(s) {
			t.Errorf("isValidPlatform(%q) = true, want false", s)
		}
	}
}

func TestIsValidRoutePattern(t *testing.T) {
	cases := []string{"/", "/foo", "/foo/bar", "/api/v1", "/a-b", "/a.b", "/a_b", "/a~b", "/foo/bar/baz", "/ABC123"}
	for _, s := range cases {
		if !isValidRoutePattern(s) {
			t.Errorf("isValidRoutePattern(%q) = false, want true", s)
		}
	}
}

func TestIsValidRoutePatternInvalid(t *testing.T) {
	cases := []string{"", "foo", "/foo//bar", "//foo", "/foo?q=1", "/foo#anchor"}
	for _, s := range cases {
		if isValidRoutePattern(s) {
			t.Errorf("isValidRoutePattern(%q) = true, want false", s)
		}
	}
}
