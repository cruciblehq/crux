package crex

import (
	"errors"
	"testing"
)

func TestWrapf(t *testing.T) {
	sentinel := errors.New("sentinel")
	wrapped := Wrapf(sentinel, "detail %d", 1)

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
