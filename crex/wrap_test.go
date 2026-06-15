package crex

import (
	"errors"
	"testing"
)

func TestWrapf(t *testing.T) {
	sentinel := errors.New("sentinel")
	cause := errors.New("cause")
	wrapped := Wrapf(sentinel, cause, "detail %d", 1)

	if !errors.Is(wrapped, sentinel) {
		t.Error("wrapped error does not match sentinel")
	}
	if !errors.Is(wrapped, cause) {
		t.Error("wrapped error does not match cause")
	}

	want := "sentinel: detail 1: cause"
	if wrapped.Error() != want {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), want)
	}
}

func TestNewf(t *testing.T) {
	sentinel := errors.New("sentinel")
	wrapped := Newf(sentinel, "detail %d", 1)

	if !errors.Is(wrapped, sentinel) {
		t.Error("wrapped error does not match sentinel")
	}

	want := "sentinel: detail 1"
	if wrapped.Error() != want {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), want)
	}
}

func TestWrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	underlying := errors.New("underlying")
	wrapped := Wrap(sentinel, underlying)

	// Test error chain preservation
	if !errors.Is(wrapped, sentinel) {
		t.Error("wrapped error does not match sentinel")
	}
	if !errors.Is(wrapped, underlying) {
		t.Error("wrapped error does not match underlying")
	}

	// Test message format
	want := "sentinel: underlying"
	if wrapped.Error() != want {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), want)
	}
}

func TestWrapf_CollapsesNestedMessages(t *testing.T) {
	outer := errors.New("outer")
	inner := errors.New("inner")
	root := errors.New("permission denied")

	innerErr := Wrapf(inner, root, "step 2")
	outerErr := Wrapf(outer, innerErr, "stage 3")

	// Only the outermost message and the foreign root render.
	want := "outer: stage 3: permission denied"
	if outerErr.Error() != want {
		t.Errorf("Error() = %q, want %q", outerErr.Error(), want)
	}

	// Every layer still participates in errors.Is matching.
	for _, target := range []error{outer, inner, root} {
		if !errors.Is(outerErr, target) {
			t.Errorf("errors.Is(outerErr, %q) = false, want true", target)
		}
	}
}

func TestTag(t *testing.T) {
	category := errors.New("category")
	specific := errors.New("specific")
	tagged := Tag(Newf(category, "woven detail"), specific)

	// Both the woven sentinel and the tag match.
	if !errors.Is(tagged, category) {
		t.Error("tagged error does not match category")
	}
	if !errors.Is(tagged, specific) {
		t.Error("tagged error does not match specific tag")
	}

	// The tag never appears in the rendered message.
	want := "category: woven detail"
	if tagged.Error() != want {
		t.Errorf("Error() = %q, want %q", tagged.Error(), want)
	}
}

func TestTag_WrapsNonWrapped(t *testing.T) {
	sentinel := errors.New("sentinel")
	tag := errors.New("tag")
	tagged := Tag(sentinel, tag)

	if !errors.Is(tagged, sentinel) {
		t.Error("tagged error does not match sentinel")
	}
	if !errors.Is(tagged, tag) {
		t.Error("tagged error does not match tag")
	}

	want := "sentinel"
	if tagged.Error() != want {
		t.Errorf("Error() = %q, want %q", tagged.Error(), want)
	}
}

func TestTag_NilReturnsNil(t *testing.T) {
	if got := Tag(nil, errors.New("tag")); got != nil {
		t.Errorf("Tag(nil, ...) = %v, want nil", got)
	}
}

func TestAt_RendersBreadcrumb(t *testing.T) {
	leaf := Newf(errors.New("invalid grant"), "empty grant")
	scoped := At(leaf, "grant", 1)
	located := At(Wrap(errors.New("invalid affordance"), scoped), "grant scope", 2)

	want := "invalid affordance: invalid grant: empty grant (at grant scope 2 > grant 1)"
	if located.Error() != want {
		t.Errorf("Error() = %q, want %q", located.Error(), want)
	}
}

func TestAt_QuotesStringPositions(t *testing.T) {
	sentinel := errors.New("invalid param")
	located := AtName(Newf(sentinel, "missing default"), "param", "name")

	want := `invalid param: missing default (at param "name")`
	if located.Error() != want {
		t.Errorf("Error() = %q, want %q", located.Error(), want)
	}
}

func TestAt_PreservesMatching(t *testing.T) {
	category := errors.New("category")
	specific := errors.New("specific")
	located := At(Tag(Newf(category, "detail"), specific), "step", 3)

	if !errors.Is(located, category) {
		t.Error("located error does not match category")
	}
	if !errors.Is(located, specific) {
		t.Error("located error does not match tag")
	}
}

func TestAt_NilReturnsNil(t *testing.T) {
	if got := At(nil, "step", 1); got != nil {
		t.Errorf("At(nil, ...) = %v, want nil", got)
	}
}
