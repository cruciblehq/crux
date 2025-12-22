package crex

import (
	"errors"
	"os"
	"testing"
)

var ErrTestSentinel = errors.New("test sentinel error")

func TestWrap(t *testing.T) {
	underlying := errors.New("underlying error")
	wrapped := Wrap(ErrTestSentinel, underlying)

	// Should preserve the error chain
	if !errors.Is(wrapped, ErrTestSentinel) {
		t.Errorf("errors.Is(wrapped, ErrTestSentinel) = false, want true")
	}

	// Should include both errors in the message
	msg := wrapped.Error()
	expectedMsg := "test sentinel error: underlying error"
	if msg != expectedMsg {
		t.Errorf("wrapped.Error() = %q, want %q", msg, expectedMsg)
	}
}

func TestWrap_WithSystemError(t *testing.T) {
	wrapped := Wrap(ErrTestSentinel, os.ErrNotExist)

	// Should preserve sentinel
	if !errors.Is(wrapped, ErrTestSentinel) {
		t.Errorf("errors.Is(wrapped, ErrTestSentinel) = false, want true")
	}

	// Should preserve underlying system error
	if !errors.Is(wrapped, os.ErrNotExist) {
		t.Errorf("errors.Is(wrapped, os.ErrNotExist) = false, want true")
	}
}

func TestWrap_WithWrappedError(t *testing.T) {
	root := errors.New("root cause")
	middle := Wrap(ErrTestSentinel, root)
	outer := Wrap(errors.New("outer sentinel"), middle)

	// Should preserve all errors in chain
	if !errors.Is(outer, ErrTestSentinel) {
		t.Errorf("errors.Is(outer, ErrTestSentinel) = false, want true")
	}
	if !errors.Is(outer, root) {
		t.Errorf("errors.Is(outer, root) = false, want true")
	}
}

func TestWrap_MessageFormat(t *testing.T) {
	tests := []struct {
		name      string
		sentinel  error
		err       error
		wantMsg   string
		wantMatch error
	}{
		{
			name:      "simple wrap",
			sentinel:  errors.New("operation failed"),
			err:       errors.New("connection timeout"),
			wantMsg:   "operation failed: connection timeout",
			wantMatch: nil,
		},
		{
			name:      "system error",
			sentinel:  errors.New("could not read file"),
			err:       os.ErrNotExist,
			wantMsg:   "could not read file: file does not exist",
			wantMatch: os.ErrNotExist,
		},
		{
			name:      "permission error",
			sentinel:  errors.New("access denied"),
			err:       os.ErrPermission,
			wantMsg:   "access denied: permission denied",
			wantMatch: os.ErrPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := Wrap(tt.sentinel, tt.err)

			msg := wrapped.Error()
			if msg != tt.wantMsg {
				t.Errorf("wrapped.Error() = %q, want %q", msg, tt.wantMsg)
			}

			if tt.wantMatch != nil && !errors.Is(wrapped, tt.wantMatch) {
				t.Errorf("errors.Is(wrapped, %v) = false, want true", tt.wantMatch)
			}
		})
	}
}
